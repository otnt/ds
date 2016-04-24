/* package main

import (
	"fmt"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/msgToPetgagMsg"
	"github.com/otnt/ds/node"
	"github.com/otnt/ds/webService"
	"github.com/otnt/ds/replication"
	"github.com/otnt/ds/message"
	"os"
	"strconv"
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
	for _, n := range infra.NodeIndexMap {
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

	//init replication
	replication.InitReplication(ring)

	go messageDispatcher()

	block := make(chan bool)
	<-block
}

func messageDispatcher() {
	//init MongoDB
	mongoSession := mongoDBintegration.EstablishSession()
	for {
		newMessage := <-infra.ReceivedBuffer
		messageKind := message.GetKind(&newMessage)
		if messageKind == "forward" {
			newPGMessage := msgToPetgagMsg.ConvertToPGMsg(&newMessage)
			replication.UpdateSelfDB(&newPGMessage, mongoSession)
			replication.AskNodesToUpdate(&newPGMessage, mongoSession)
			go replication.WaitForAcks()
			replication.RespondToClient()
			// At the secondary
		} else if messageKind == "replication" {
			newPGMessage := msgToPetgagMsg.ConvertToPGMsg(&newMessage)
			replication.UpdateSelfDB(&newPGMessage, mongoSessions)
			replication.SendAcks(&newPGMessage)
			// Acks processing at the primary
		} else if messageKind == "acknowledgement" {
			replication.ProcessAcks()
		} else if messageKind == webService.KIND_FORWARD {
			webService.ForwardChan <- &newMessage
		} else if messageKind == webService.KIND_FETCH {
			webService.FetchChan <- &newMessage
		} else if messageKind == webService.KIND_FORWARD_ACK {
			webService.ForwardAckChan <- &newMessage
		} else if messageKind == webService.KIND_FETCH_ACK {
			webService.FetchAckChan <- &newMessage
		} else if messageKind == webService.KIND_COMMENT {
			webService.CommentChan <- &newMessage
		} else if messageKind == webService.KIND_COMMENT_ACK {
			webService.CommentAckChan <- &newMessage
		} else if messageKind == webService.KIND_UP_VOTE{
			webService.UpVoteChan<- &newMessage
		} else if messageKind == webService.KIND_UP_VOTE_ACK {
			webService.UpVoteAckChan <- &newMessage
		} else if messageKind == webService.KIND_DOWN_VOTE{
			webService.DownVoteChan<- &newMessage
		} else if messageKind == webService.KIND_DOWN_VOTE_ACK {
			webService.DownVoteAckChan <- &newMessage
		}
	}
}*/
