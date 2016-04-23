package replication

//import (
//	"bufio"
//	"bytes"
//	"encoding/json"
//	"fmt"
//	ch "github.com/otnt/ds/consistentHashing"
//	"github.com/otnt/ds/infra"
//	"github.com/otnt/ds/message"
//	"github.com/otnt/ds/mongoDbintegration"
//	"github.com/otnt/ds/msgToPetgagMsg"
//	"github.com/otnt/ds/node"
//	//"github.com/otnt/ds/petGagMessage"
//	"github.com/otnt/ds/petgagData"
//	//"github.com/otnt/ds/replication"
//	//"github.com/otnt/ds/webService"
//	"os"
//	//"strconv"
//	"strings"
//	"time"
//)
//
//func replicationTest() {
//	//init infra
//	if len(os.Args) != 3 {
//		fmt.Fprintf(os.Stderr, "Usage: %s Hostname Port\n", os.Args[0])
//		os.Exit(1)
//	}
//	localHost := os.Args[1]
//	infra.InitNetwork(localHost)
//
//	time.Sleep(500)
//	fmt.Println("You can start sending messages")
//	fmt.Println("<dest> <kind>")
//	//init consistent hashing
//	ring := ch.NewRing()
//	for _, n := range infra.NodeIndexMap {
//		nn := node.Node(*n)
//		ring.AddSync(&nn)
//	}
//
//	//init web service
//	/*port, err := strconv.Atoi(os.Args[2])
//	if err != nil {
//		panic(err)
//	}*/
//
//	//init replication
//	InitReplication(ring)
//
//	go senderApp()
//	go messageDispatcher()
//
//	block := make(chan bool)
//	<-block
//}
//
//func senderApp() {
//	reader := bufio.NewReader(os.Stdin)
//	belongsTo := ""
//	dbOp := "Insert"
//	imageURL := "https://www.petfinder.com/wp-content/uploads/2012/11/99233806-bringing-home-new-cat-632x475.jpg"
//	var commtUN string = ""
//	var commtComment string = ""
//	upVote := 0
//	downVote := 0
//	userName := "prathi"
//	userID := "12345"
//	objID := "nil"
//
//	/* Construct the comment */
//	comment := petgagData.NewComment(commtComment, commtUN)
//
//	/* Construct petgagData */
//	newPGData := petgagData.NewPetGagData(dbOp, comment, upVote, downVote, userName, userID, objID, imageURL, belongsTo)
//
//	/* Encode PGData struct into a string */
//	var buf bytes.Buffer
//	json.NewEncoder(&buf).Encode(newPGData)
//	data := buf.String()
//
//	//for {
//	input, _ := reader.ReadString('\n')
//	input = input[:len(input)-1]
//	//s_input := strings.SplitN(input, " ", 3)
//	s_input := strings.SplitN(input, " ", 2)
//	dest := s_input[0]
//	kind := s_input[1]
//	//data := s_input[2]
//
//	fmt.Println("The destination is:", dest)
//	fmt.Println("The kind is: ", kind)
//	fmt.Println("The encoded data is: ", data)
//	infra.SendUnicast(dest, data, kind)
//	//}
//}
//
//func messageDispatcher() {
//	//init MongoDB
//	fmt.Println("Inside messageDispatcher of main")
//	mongoSession := mongoDBintegration.EstablishSession()
//	for {
//		select {
//		case newMessage := <-infra.ReceivedBuffer:
//			messageKind := message.GetKind(&newMessage)
//
//			fmt.Println("Received New Message!!")
//			fmt.Println("The New Message is ", newMessage)
//
//			if messageKind == "forward" {
//				fmt.Println("I am the primary")
//				newPGMessage := msgToPetgagMsg.ConvertToPGMsg(&newMessage)
//
//				localNode := infra.GetLocalNode()
//				newPGMessage.PGData.BelongsTo = localNode.Hostname
//
//				UpdateSelfDB(&newPGMessage, mongoSession)
//
//				fmt.Println("Asking nodes to update")
//
//				AskNodesToUpdate(&newPGMessage, mongoSession)
//
//				fmt.Println("Waiting for acknowledgements to come in")
//				go func() {
//					var acksObtained int = 0
//					for {
//						acksObtained = NumAcks
//						if acksObtained == ReplicationFactor {
//							fmt.Println("acks obtained = ", NumAcks)
//							break
//						}
//					}
//					RespondToClient()
//				}()
//
//				//go replication.WaitForAcks()
//
//			} else if messageKind == "replication" { /* At the secondary */
//
//				fmt.Println("I am the secondary")
//
//				localNode := infra.GetLocalNode()
//				newPGMessage := msgToPetgagMsg.ConvertToPGMsg(&newMessage)
//				newPGMessage.PGData.BelongsTo = localNode.Hostname + " - replication"
//
//				UpdateSelfDB(&newPGMessage, mongoSession)
//
//				fmt.Println("Updated secondary DB. Sending acks")
//				SendAcks(&newPGMessage)
//
//			} else if messageKind == "acknowledgement" { /* Acks processing at the primary */
//				fmt.Println("Received an ack")
//				ProcessAcks()
//
//			} else {
//				fmt.Println("Enter the correct kind of message")
//			}
//
//		case <-time.After(time.Millisecond * 1):
//			continue
//		}
//	}
//}
