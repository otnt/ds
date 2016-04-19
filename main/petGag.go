package main

import (
<<<<<<< HEAD
	"fmt"
=======
	"github.com/otnt/ds/webService"
	//"github.com/otnt/ds/replication"
>>>>>>> 1345e4781de4a6cf0e1dec162bc250490e00228e
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/node"
	"github.com/otnt/ds/replication"
	"github.com/otnt/ds/webService"
	"os"
	"strconv"
<<<<<<< HEAD
	"time"
=======
	"github.com/otnt/ds/message"
>>>>>>> 1345e4781de4a6cf0e1dec162bc250490e00228e
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
	for _, n := range nodes.Servers {
		nn := node.Node(n)
=======
	for _,n:= range infra.NodeIndexMap{
		nn := node.Node(*n)
>>>>>>> 1345e4781de4a6cf0e1dec162bc250490e00228e
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
		select {
		case newMessage := <-infra.ReceivedBuffer:
			messageKind := message.GetKind(&newMessage)

<<<<<<< HEAD
		if messageKind == "forward" { /* At the primary */
			replication.UpdateSelfDB(&newMessage)
			replication.AskNodesToUpdate(&newMessage)
			go replication.WaitForAcks()
			replication.RespondToClient()

		} else if messageKind == "forward" { /* At the secondary */
			replication.UpdateSelfDB(&newMessage)
			replication.SendAcks(&newMessage)
		} else if messageKind == "acknowledgement" { /* Acks processing at the primary */
			replication.ProcessAcks()
=======
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

		case <-time.After(time.Millisecond * 1):
			continue
>>>>>>> 1345e4781de4a6cf0e1dec162bc250490e00228e
		}
	}
}
