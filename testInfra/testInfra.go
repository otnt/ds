package main

import (
	"fmt"
	"strings"
	"os"
	"bufio"
	"time"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/message"
)	

func main() {
        if len(os.Args) != 2 {
                fmt.Fprintf(os.Stderr, "Usage: %s Hostname\n", os.Args[0])
                os.Exit(1)
        }
        localHost := os.Args[1]
	infra.InitNetwork(localHost)	
        /* Test code for the send and listener threads */
        
        time.Sleep(500)
	fmt.Println("You can start sending messages")
	fmt.Println("<dest> <kind> <data to send>")
	go receiverThread()
	go senderApp()

	/* DON'T CARE ABOUT THIS -- Dummy code to make the application block here and not exit */
	dummyChan := make (chan string)
	dummyString := <- dummyChan
	fmt.Println(dummyString)
}

func receiverThread() {
	for {
		newMessage := infra.CheckIncomingMessages()
		fmt.Printf("TestInfra [Kind: %s] %s: %s\n", message.GetKind(&newMessage), message.GetSrc(&newMessage), message.GetData(&newMessage))
	}
}



func senderApp() {
	reader:=bufio.NewReader(os.Stdin)
        for {
		input, _ := reader.ReadString('\n')
		input = input[:len(input)-1]
		s_input := strings.SplitN(input, " ", 3)
		dest:= s_input[0]
		kind := s_input[1]
		data := s_input[2]
		infra.SendUnicast(dest, data, kind)
        }
}
