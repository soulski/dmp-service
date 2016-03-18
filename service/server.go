package service

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/soulski/dmp-service/dmp"
	"github.com/streadway/amqp"
)

const (
	DMP_PORT     = 8080
	SERVICE_PORT = 80

	JSON_CONTENT_TYPE = "application/json"
)

type action struct {
	url    string
	action func(w http.ResponseWriter, r *http.Request)
}

var serverUrlAction = []*action{
	{url: "/dmp", action: actionDMP},
}

type Server struct {
	ns string
}

func CreateServer(ns string) *Server {
	return &Server{
		ns: ns,
	}
}

func (s *Server) Start() {
	for _, action := range serverUrlAction {
		http.HandleFunc(action.url, action.action)
	}

	service := &dmp.Service{
		Namespace:    s.ns,
		ContactPoint: fmt.Sprintf("http://127.0.0.1:%d/dmp", SERVICE_PORT),
	}

	if result, err := dmp.RegisterService(service); !result {
		if err != nil {
			panic("Fail to register DMP \n Eror : " + err.Error())
		}
		panic("Fail to register DMP")
	}

	fmt.Println("[Service][Info]Service is started...")

	go startMsgQ()
	http.ListenAndServe(fmt.Sprintf(":%d", SERVICE_PORT), nil)
}

func actionDMP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "DMP receiver serve on HTTP Method 'PUT'")
		return
	}

	ioutil.ReadAll(r.Body)
	//	fmt.Printf("[Service][Info]Receive msg : %s \n", string(bodyBytes))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Thank for message."))
}

func startMsgQ() {
	conn, err := amqp.Dial("amqp://guest:guest@173.17.0.6:5672")
	if err != nil {
		fmt.Println("Error : ", err.Error())
		return
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println("Error : ", err.Error())
		return
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"rpc_queue", // name
		false,       // durable
		false,       // delete when usused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		fmt.Println("Error : ", err.Error())
		return
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		fmt.Println("Error : ", err.Error())
		return
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		fmt.Println("Error : ", err.Error())
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			n := string(d.Body)
			fmt.Printf("Recv : %s\n", n)

			response := "Hi There!"

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				})
			if err != nil {
				fmt.Println("Error : ", err.Error())
				continue
			}

			d.Ack(false)
		}
	}()

	fmt.Println(" [*] Awaiting RPC requests")
	<-forever
}
