package main

import (
	"fmt"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/infra"
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
	ring := ch.NewRing()
	for _, n := range nodes.Servers {
		nn := node.Node(n)
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
}

func messageDispatcher() {
	for {
		newMessage := infra.CheckIncomingMessages()
		messageKind := message.GetKind(&newMessage)

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
		}
	}
}
