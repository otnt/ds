package main

import (
	"fmt"
<<<<<<< HEAD
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/node"
	"github.com/otnt/ds/replication"
	"github.com/otnt/ds/webService"
	//"github.com/otnt/ds/replication"
	"github.com/otnt/ds/message"
	"os"
	"strconv"
=======
	"github.com/otnt/ds/webService"
	//"github.com/otnt/ds/replication"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/node"
	//"github.com/otnt/ds/replication"
	"os"
	"strconv"
	"github.com/otnt/ds/message"
>>>>>>> 4fd99c42d483b69cab174894d2571b3a4dd23f93
	"time"
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
<<<<<<< HEAD
	for _, n := range infra.NodeIndexMap {
=======
	for _,n:= range infra.NodeIndexMap{
>>>>>>> 4fd99c42d483b69cab174894d2571b3a4dd23f93
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
	//init MongoDB
	mongoSession := replication.InitMongoDB()
	for {
		select {
		case newMessage := <-infra.ReceivedBuffer:
			messageKind := message.GetKind(&newMessage)

<<<<<<< HEAD
			if messageKind == "forward" {
				replication.UpdateSelfDB(&newMessage, mongoSession)
				replication.AskNodesToUpdate(&newMessage, mongoSession)
				go replication.WaitForAcks()
				replication.RespondToClient()
			} else if messageKind == "replication" { /* At the secondary */
				replication.UpdateSelfDB(&newMessage, mongoSessions)
				replication.SendAcks(&newMessage)
			} else if messageKind == "acknowledgement" { /* Acks processing at the primary */
				replication.ProcessAcks()
=======
			if messageKind == "replication" {
				//replication.NameOfFunction(&newMessage)
>>>>>>> 4fd99c42d483b69cab174894d2571b3a4dd23f93
			} else if messageKind == webService.KIND_FORWARD {
				webService.ForwardChan <- &newMessage
			} else if messageKind == webService.KIND_FETCH {
				webService.FetchChan <- &newMessage
			} else if messageKind == webService.KIND_FORWARD_ACK {
				webService.ForwardAckChan <- &newMessage
			} else if messageKind == webService.KIND_FETCH_ACK {
				webService.FetchAckChan <- &newMessage
			}

		case <-time.After(time.Millisecond * 1):
			continue
		}
	}
}
