package config

import (
	"gopkg.in/yaml.v2"
	"github.com/otnt/ds/node"
	"io/ioutil"
)

type Nodes struct {
	List []node.Node
}

const (
	BootstrapNodesFile = "/home/ubuntu/work/src/github.com/otnt/ds/config/nodes.yml"
)

type YamlConfig struct {
	Servers []node.Node
}

func (c *YamlConfig) ParseYaml(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic (err)
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		panic (err)
	}

	UpdatedServers := make([]node.Node, len(c.Servers))
	for index, n := range(c.Servers) {
		UpdatedServers[index] = *node.NewNode(n.Hostname, n.Ip, n.Port, 2)
	}
	c.Servers = UpdatedServers
}

/*
func BootstrapNodes() Nodes{
	data, err := ioutil.ReadFile(BootstrapNodesFile)
	if err != nil {
		panic(err)
	}

	fmt.Println(data)

	var nodes,tmp Nodes
	err = yaml.Unmarshal(data, &tmp)
	if err != nil {
		panic(err)
	}
	fmt.Println(tmp)

	nodes.List = make([]node.Node, len(tmp.List))
	for index, n := range tmp.List {
		nodes.List[index] = *node.NewNode(n.Hostname, n.Ip, n.Port, 2)
	}

	return nodes
}
*/
