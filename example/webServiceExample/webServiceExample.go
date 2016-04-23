package main

import (
	"github.com/otnt/ds/webService"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
	"fmt"
	"os"
	"github.com/otnt/ds/node"
	"time"
	"strconv"
	"github.com/otnt/ds/message"
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
	ws:= webService.WebService{Port: port}
	ws.Run(ring)

	//incoming message dispatcher
	go func() {
		for {
			select {
				case newMessage := <-infra.ReceivedBuffer:  //infra.CheckIncomingMessages()
				messageKind := message.GetKind(&newMessage)
				fmt.Println("receive msg kind " + messageKind)
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

			case <-time.After(time.Millisecond * 1):
				continue
			}
		}
	}()

	block := make(chan bool)
	<-block
}
