package infra

import (
	"fmt"
	"io/ioutil"
	"os"
	"net"
	"strings"
	"github.com/melhalab/node"
	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	Servers []node.Node
}

func (c *YamlConfig) ParseYaml(fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	checkError(err)
	err = yaml.Unmarshal(data, c)
	fmt.Println("Printing yamlConfig")
	fmt.Printf("%+v\n", c)
	fmt.Println("Done Printing yamlConfig")
	return err
}

/* Hashmap for nodes on the network */
var nodeIndexMap map[string]*node.Node

/* Hashmap for connections to servers */
var connectionMap map[string]*net.TCPConn

func listenerThread(conn *net.TCPConn) {
	message := make([]byte, 128)
	defer conn.Close()  // close connection at exit
	for {
		read_len, err := conn.Read(message)
		if err != nil {
			fmt.Println(err)
			break
		}
		if read_len == 0 {
			/* Connection closed by the client, exit thread */
			break
		} else {
			fmt.Printf("New message received from %s: %s\n", conn.RemoteAddr(), string(message))
		}
		/* Clear message for next read */
		message = make([]byte, 128)
	}
}

func SendUnicast(dest string, message string) {
	conn := connectionMap[dest]
	_, err := conn.Write([]byte(message))
	checkError(err)
}

func connectToNode(node *node.Node, localHost string) {
	/* Get the remote node TCP address */
	service := fmt.Sprintf("%s:%d", node.Ip, node.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	/* Connect to Server */
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	/* Send a connection message to identify self */
	connectionMessage := fmt.Sprintf("HELO MESSAGE FROM %s CONNECT", localHost)
	_, err = conn.Write([]byte(connectionMessage))
	checkError(err)

	/* Save the net.Conn pointer to map using remote server's Uuid */
	connectionMap[node.Hostname] = conn
	go listenerThread(conn)
}

func connectToOtherServers(pYamlConfig *YamlConfig, localHost string) {
	for _, each := range pYamlConfig.Servers {
		if each.Hostname > localHost {
			remoteNode := nodeIndexMap[each.Hostname]
			connectToNode(remoteNode, localHost)
			fmt.Println("Connected to", remoteNode.Hostname, " at ", remoteNode.Uuid)
		}
	}
}

func acceptConnectionsFromOtherServers(pYamlConfig *YamlConfig, localHost string) {
	localNode := nodeIndexMap[localHost]

	/* Counter for how many servers will connect to me */
	connectionCount := 0
	/* Every server before localHost in the list will attempt to connect */
	for _, each := range pYamlConfig.Servers {
                if each.Hostname < localHost {
			connectionCount++
                }
        }
	service := fmt.Sprintf(":%d", localNode.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	fmt.Println("*** Accepting Connections on", tcpAddr)
	checkError(err)
	connMessage := make([]byte, 128)
	for connectionCount > 0 {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		/* Read the connection message from the remote server */
		read_len, err := conn.Read(connMessage)
		if read_len == 0 {
			fmt.Println("Connection to remote host failed *1")
			break
		}
		fmt.Println("Connection message received: ", string(connMessage))
		s_connMessage := strings.SplitN(string(connMessage), " ", 5)
		if s_connMessage[0] != "HELO" {
			fmt.Println("Connection failed - no HELO message received")
			break
		}
		remoteHost := s_connMessage[3]
		remoteNode := nodeIndexMap[remoteHost]
		connectionMap[remoteNode.Hostname] = conn.(*net.TCPConn)
		//go handleClient(conn)
		fmt.Printf("Connected to [%s] at %s\n", remoteNode.Hostname, conn.RemoteAddr())
		go listenerThread(conn.(*net.TCPConn))
		connectionCount--
	}
	fmt.Println("*** Done Accepting Connections ***")
}

func InitNetwork(localHost string) {
	nodeIndexMap = map[string]*node.Node{}
	connectionMap = map[string]*net.TCPConn{}
	vnodeNum := 2
	var yamlConfig YamlConfig
	err := yamlConfig.ParseYaml("nodes.yml")
	checkError(err)
	for _, each := range yamlConfig.Servers {
		/* Build the nodeIndexMap hashmap 
		if each.Hostname = localHost {
			localHostIndex = index
		} */
		nodeIndexMap[each.Hostname] = node.NewNode(each.Hostname, each.Ip, each.Port, vnodeNum)
	}

	for key, value := range nodeIndexMap {
	    fmt.Println("Key:", key, "Value:", value)
	}

	localNode := nodeIndexMap[localHost]
	//localNode = yamlConfig.Servers[localHostIndex]
	fmt.Printf("Local Host is [%s] at %s:%d\n", localNode.Hostname, localNode.Ip, localNode.Port)
	fmt.Printf("Keys are: %+v\n", localNode.Keys)
	acceptConnectionsFromOtherServers(&yamlConfig, localHost)
	connectToOtherServers(&yamlConfig, localHost)
	fmt.Println("***************************************************")
	fmt.Println("****** Done all conections and back to main *******")
	fmt.Println("***************************************************")
}

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
        os.Exit(1)
    }
}
