package main

import (
	"fmt"
	"github.com/otnt/ds/webService"
	//"github.com/otnt/ds/replication"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/node"
	//"github.com/otnt/ds/replication"
	"os"
	"strconv"
	"github.com/otnt/ds/message"
	//"time"
)

func main() {
	//init infra
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s Hostname Port\n", os.Args[0])
		os.Exit(1)
	}
	localHost := os.Args[1]
	infra.InitNetwork(localHost)

	//init consistent hashing
	ring := ch.NewRing()
	for _,n:= range infra.NodeIndexMap{
		nn := node.Node(*n)
		ring.AddSync(&nn)
	}

	//init web service
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	ws := webService.WebService{Port: port}
	ws.Run(ring)
	go messageDispatcher()

	block := make(chan bool)
	<-block
}

func messageDispatcher() {
	for {
		newMessage := <-infra.ReceivedBuffer
		messageKind := message.GetKind(&newMessage)

		if messageKind == "replication" {
			//replication.NameOfFunction(&newMessage)
		} else if messageKind == webService.KIND_FORWARD {
			webService.ForwardChan <- &newMessage
		} else if messageKind == webService.KIND_FETCH {
			webService.FetchChan <- &newMessage
		} else if messageKind == webService.KIND_FORWARD_ACK {
			webService.ForwardAckChan <- &newMessage
		} else if messageKind == webService.KIND_FETCH_ACK {
			webService.FetchAckChan <- &newMessage
		}
	}
}
