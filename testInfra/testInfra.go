package main

import (
	"fmt"
	"strings"
	"os"
	"bufio"
	"time"
	"github.com/otnt/ds/infra"
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
	reader:=bufio.NewReader(os.Stdin)
        for {
		input, _ := reader.ReadString('\n')
		s_input := strings.SplitN(input, " ", 2)
		dest:= s_input[0]
		message:= s_input[1]
		infra.SendUnicast(dest, message)
        }
}
