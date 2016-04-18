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
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	ws:= webService.WebService{Port: port}
	ws.Run(ring)
}
