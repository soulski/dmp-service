package main

import (
	"os"

	"github.com/soulski/dmp-service/service"
)

func main() {
	var sType, ns string

	if len(os.Args) == 3 {
		ns = os.Args[2]
		sType = os.Args[1]
	} else {
		ns = os.Getenv("SERVICE_NS")
		sType = os.Getenv("SERVICE_TYPE")
	}

	switch sType {
	case "client":
		service.CreateClient(ns).Start()
	case "server":
		service.CreateServer(ns).Start()
	default:
		panic("Unknow " + sType + " type service")
	}
}
