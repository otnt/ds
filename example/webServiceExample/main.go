package main

import (
	"fmt"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/node"
	"github.com/otnt/ds/replication"
	"github.com/otnt/ds/webService"
	"os"
	"strconv"
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
	time.Sleep(500)

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

	//incoming message dispatcher
	go func() {
		for {
			select {
			case newMessage := <-infra.ReceivedBuffer: //infra.CheckIncomingMessages()
				messageKind := message.GetKind(&newMessage)
				fmt.Println("receive msg kind " + messageKind)
				if messageKind == replication.KIND_REPLICATION {
					fmt.Println("Received message to replicate")
					replication.ReplChan <- &newMessage
				} else if messageKind == replication.KIND_REPLICN_ACK {
					replication.AckChan <- &newMessage
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
				} else if messageKind == webService.KIND_UP_VOTE {
					webService.UpVoteChan <- &newMessage
				} else if messageKind == webService.KIND_UP_VOTE_ACK {
					webService.UpVoteAckChan <- &newMessage
				} else if messageKind == webService.KIND_DOWN_VOTE {
					webService.DownVoteChan <- &newMessage
				} else if messageKind == webService.KIND_DOWN_VOTE_ACK {
					webService.DownVoteAckChan <- &newMessage
				} else if messageKind == replication.KIND_TRANSFER {
					replication.TransferChan <- &newMessage
				}

			case <-time.After(time.Millisecond * 1):
				continue
			}
		}
	}()

	block := make(chan bool)
	<-block
}

/*
package main

import (
	//	"bufio"
	"fmt"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/node"
	"github.com/otnt/ds/replication"
	"os"
	//	"strings"
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

	time.Sleep(500)
	fmt.Println("You can start sending messages")
	//fmt.Println("<dest> <kind>")
	//init consistent hashing
	ring := ch.NewRing()
	for _, n := range infra.NodeIndexMap {
		nn := node.Node(*n)
		ring.AddSync(&nn)
	}


	//init replication
	replication.InitReplication(ring)
	go senderApp()
	go messageDispatcher()

	block := make(chan bool)
	<-block
}

func senderApp() {
	//reader := bufio.NewReader(os.Stdin)
	if infra.GetLocalNode().Hostname == "Alice" {
		replication.SendCollnToNewPrimary("Charlie")
	} else if infra.GetLocalNode().Hostname == "Charlie" {
		replication.MergeCollections("Bob")
	}

}

func messageDispatcher() {
	//init MongoDB
	fmt.Println("Inside messageDispatcher of main")
	for {
		select {
		case newMessage := <-infra.ReceivedBuffer:
			messageKind := message.GetKind(&newMessage)

			fmt.Println("Received New Message!!")
			fmt.Println("The New Message is ", newMessage)

			if messageKind == replication.KIND_TRANSFER {
				time.Sleep(100 * time.Millisecond)
				replication.AddNewCollection(&newMessage)
			}

		case <-time.After(time.Millisecond * 1):
			continue
		}
	}
}*/
