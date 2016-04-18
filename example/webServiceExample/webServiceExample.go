package main

import (
	"github.com/otnt/ds/webService"
	ch "github.com/otnt/ds/consistentHashing"
	config "github.com/otnt/ds/config"
	"github.com/otnt/ds/infra"
	"fmt"
	"os"
	"github.com/otnt/ds/node"
	"time"
)

func main() {
	//init infra
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s Hostname\n", os.Args[0])
		os.Exit(1)
	}
	localHost := os.Args[1]
	infra.InitNetwork(localHost)
	time.Sleep(500)

	//init consistent hashing
	//nodes := config.BootstrapNodes()
	var nodes config.YamlConfig
	nodes.ParseYaml(config.BootstrapNodesFile)
	ring := ch.NewRing()
	for _,n:= range nodes.Servers{
		nn := node.Node(n)
		ring.AddSync(&nn)
	}

	//init web service
	ws:= webService.WebService{Port:8081}
	ws.Run(ring)
}
