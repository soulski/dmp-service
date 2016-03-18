package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/soulski/dmp-service/dmp"
	"github.com/streadway/amqp"
)

var mappingUrlAction = []*action{
	{url: "/dmp", action: actionDMP},
	{url: "/message", action: actionSendMsg},
	{url: "/benchmark/point2point", action: actionBenchmarkPoint2Point},
}

type Client struct {
	ns string
}

func CreateClient(ns string) *Client {
	return &Client{
		ns: ns,
	}
}

func (c *Client) Start() {
	for _, action := range mappingUrlAction {
		http.HandleFunc(action.url, action.action)
	}

	service := &dmp.Service{
		Namespace:    c.ns,
		ContactPoint: fmt.Sprintf("http://127.0.0.1:%d/dmp", SERVICE_PORT),
	}

	if result, err := dmp.RegisterService(service); !result {
		if err != nil {
			panic("Fail to register DMP \n Eror : " + err.Error())
		}
		panic("Fail to register DMP")
	}

	fmt.Println("[Service][Info]Service is started...")

	http.ListenAndServe(fmt.Sprintf(":%d", SERVICE_PORT), nil)
}

func actionSendMsg(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "DMP receiver serve on HTTP Method 'PUT'")
		return
	}

	var msg dmp.Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&msg); err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Error : Invalid req format")
		return
	}

	resStr, err := dmp.SendMsg(&msg)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Error : While sending message to dmp")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resStr))
}

func actionBenchmarkPoint2Point(w http.ResponseWriter, r *http.Request) {
	reqNo := 10

	msgs := make([]string, 0, reqNo)
	for index := 0; index < reqNo; index++ {
		msgs = append(msgs, RandString(rand.Intn(100)))
	}

	DMPResult := Benchmark(reqNo, func(loopNo int) bool {
		_, err := dmp.SendMsg(&dmp.Message{
			Type:      "req-res",
			Namespace: "ns2",
			Body:      msgs[loopNo],
		})

		if err != nil {
			return false
		}

		return true
	})

	var target string
	var directResult *Result

	members, err := dmp.GetMembers("ns2")
	if err != nil {
		fmt.Println("Error : " + err.Error())
		goto SKIP_DIRECT
	}

	target = members.Members[0]

	directResult = Benchmark(reqNo, func(loopNo int) bool {
		client := &http.Client{}
		req, err := http.NewRequest(
			"PUT",
			fmt.Sprintf("http://%s:80", target),
			bytes.NewReader([]byte(msgs[loopNo])),
		)

		if err != nil {
			return false
		}

		req.Header.Set("Content-Type", JSON_CONTENT_TYPE)

		res, err := client.Do(req)
		if err != nil {
			return false
		}
		defer res.Body.Close()

		ioutil.ReadAll(res.Body)
		return true
	})

SKIP_DIRECT:

	var conn *amqp.Connection
	var ch *amqp.Channel
	var q amqp.Queue
	var rabbitResult *Result

	conn, err = amqp.Dial("amqp://guest:guest@173.17.0.6:5672")
	if err != nil {
		fmt.Println("Error : ", err.Error())
		goto SKIP_RABBITMQ
	}
	defer conn.Close()

	rabbitResult = Benchmark(reqNo, func(loopNo int) bool {
		ch, err = conn.Channel()
		if err != nil {
			fmt.Println("Error : ", err.Error())
			return false
		}

		defer ch.Close()
		q, err = ch.QueueDeclare(
			"test", //name
			false,  //durable
			false,  //delete when unused
			false,  //exclusive
			false,  //no-wait
			nil,    //arguments
		)

		if err != nil {
			fmt.Println("Error : ", err.Error())
			return false
		}

		msgCh, err := ch.Consume(
			q.Name, //queue
			"",     //consumer
			true,   //auto-ack
			false,  //exclusive
			false,  //no-local
			false,  //no-wait
			nil,    //args
		)
		if err != nil {
			fmt.Println("Error : ", err.Error())
			return false
		}

		corrId := randomString(32)

		err = ch.Publish(
			"",          // exchange
			"rpc_queue", // routing key
			false,       // mandatory
			false,       // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrId,
				ReplyTo:       q.Name,
				Body:          []byte(msgs[loopNo]),
			})

		if err != nil {
			fmt.Println("Error : ", err.Error())
			return false
		}

		for d := range msgCh {
			if corrId == d.CorrelationId {
				res := string(d.Body)
				fmt.Printf("[Rabbitmq] Recv : %s\n", res)
				if err != nil {
					fmt.Println("Error : ", err.Error())
					return false
				}
				break
			}
		}

		return true
	})

SKIP_RABBITMQ:

	writeResult := func(w http.ResponseWriter, name string, r *Result) {
		if r == nil {
			return
		}

		resFormat := `Benchmark of %s result
Attempt to send %d times
Success %d times
Average time is %d nanosecond
Average time is %d millisecond
Max time is %d nanosecond
Max time is %d millisecond
Min time is %d nanosecond
Min time is %d millisecond
	`

		if r != nil {
			fmt.Fprintf(
				w,
				resFormat,
				name,
				r.TaskNo,
				r.SuccessNo,
				r.AverageTime,
				r.AverageTime/1000,
				r.MaxTime,
				r.MaxTime/1000,
				r.MinTime,
				r.MinTime/1000,
			)
		}

	}

	w.WriteHeader(http.StatusOK)
	writeResult(w, "DMP", DMPResult)
	writeResult(w, "Direct", directResult)
	writeResult(w, "Rabbit", rabbitResult)
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
