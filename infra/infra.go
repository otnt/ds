package infra

import (
	"fmt"
	"io/ioutil"
	"os"
	"net"
	"strings"
	"time"
	"github.com/otnt/ds/clockService"
	"github.com/otnt/ds/node"
	"github.com/otnt/ds/message"
	"gopkg.in/yaml.v2"
)

type YamlConfig struct {
	Servers []node.Node
}

func (c *YamlConfig) ParseYaml(fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	checkError(err)
	err = yaml.Unmarshal(data, c)
	return err
}

/* Hashmap for nodes on the network  */
var NodeIndexMap map[string]*node.Node

/* Hashmap for connections to servers */
var connectionMap map[string]*net.TCPConn

/* Node for the local host */
var LocalNode *node.Node

/* Logical clock for infra */
var NodeLogicalClock clockService.LogicalClock

var ReceivedBuffer chan message.Message

func listenerThread(conn *net.TCPConn) {
	readFromSocket := make([]byte, 4096)
	defer conn.Close()  // close connection at exit
	for {
		read_len, err := conn.Read(readFromSocket)
		if err != nil {
			fmt.Println(err)
			break
		}
		if read_len == 0 {
			/* Connection closed by the client, exit thread */
			break
		} else {
			var rcvMessage message.Message
			err := message.Unmarshal(readFromSocket[:read_len], &rcvMessage)
			checkError(err)
	
			/* Update the timestamp according to received message */
			NodeLogicalClock.UpdateTimestamp(&clockService.LogicalTimestamp{rcvMessage.Timestamp})
			
			/* Add message to ReceivedBuffer */
			go func() { ReceivedBuffer <- rcvMessage }()
		}
		/* Clear message for next read */
		readFromSocket = make([]byte, 4096)
	}
	
}

func SendUnicast(dest string, data string, kind string) {

	/* Request a new timestamp for the outgoing message */
	NodeLogicalClock.NewTimestamp()

	/* Create a new Message with local timestamp and hostname and other parameters from argument list */
	sendMessage := message.NewMessage(NodeLogicalClock.GetCurrentTimestamp(), LocalNode.Hostname, dest, data, kind)

	if dest == LocalNode.Hostname {
		/* Destination is local node, just put message in ReceivedBuffer without going to network */
		go func() {
			ReceivedBuffer <- sendMessage
			NodeLogicalClock.NewTimestamp()
		}()
		return
	}

	conn, ok := connectionMap[dest]
	if ok {
		_, err := conn.Write(message.Marshal(&sendMessage))
		if err != nil {
			/* Connection lost with the server, remove the server from the connection map and return */
			fmt.Println("Error: Connection is lost with", dest)
			delete(connectionMap, dest)
		}
	} else {
		fmt.Println("Error: Destination [",dest,"] unknown")
	}
}

func connectToNode(node *node.Node) int {
	/* Get the remote node TCP address */
	service := fmt.Sprintf("%s:%d", node.Ip, node.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error [net.ResolveTCPAddr]: %s\n", err.Error())
		return -1;
	}
	
	/* Try to connect to the remote, sleep 5 seconds between retries */
	var conn *net.TCPConn
	for {
	
		/* Connect to Server */
		conn, err = net.DialTCP("tcp", nil, tcpAddr)
		if err == nil {
			break
		} else {
			fmt.Fprintf(os.Stderr, "Error [net.DialTCP]: %s -- Retrying in 5 seconds\n", err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
	}
	
	/* Send a connection message to identify self */
	connectionMessage := fmt.Sprintf("HELO MESSAGE FROM %s CONNECT", LocalNode.Hostname)
	_, err = conn.Write([]byte(connectionMessage))
	checkError(err)
	
	/* Save the net.Conn pointer to map using remote server's Uuid */
	connectionMap[node.Hostname] = conn
	go listenerThread(conn)
	return 1
}

func connectToOtherServers(pYamlConfig *YamlConfig) {
	for _, each := range pYamlConfig.Servers {
		if each.Hostname > LocalNode.Hostname {
			remoteNode := NodeIndexMap[each.Hostname]
			if connectToNode(remoteNode) > 0 {
				fmt.Println("Connected to", remoteNode.Hostname, "at", remoteNode.Uuid)
			} else {
				fmt.Println("Failed to connect to", remoteNode.Hostname, "at", remoteNode.Uuid)
				os.Exit(1)
			}
		}
	}
}

func acceptConnectionsFromOtherServers(pYamlConfig *YamlConfig) {
	/* Counter for how many servers will connect to me */
	connectionCount := 0
	/* Every server before localHost in the list will attempt to connect */
	for _, each := range pYamlConfig.Servers {
                if each.Hostname < LocalNode.Hostname {
			connectionCount++
                }
        }
	service := fmt.Sprintf(":%d", LocalNode.Port)
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
		remoteNode := NodeIndexMap[remoteHost]
		connectionMap[remoteNode.Hostname] = conn.(*net.TCPConn)
		//go handleClient(conn)
		fmt.Printf("Connected to [%s] at %s\n", remoteNode.Hostname, conn.RemoteAddr())
		go listenerThread(conn.(*net.TCPConn))
		connectionCount--
	}
	fmt.Println("*** Done Accepting Connections ***")
}

func InitNetwork(localHost string) {
	NodeLogicalClock = clockService.NewLogicalClock()
	NodeIndexMap = map[string]*node.Node{}
	connectionMap = map[string]*net.TCPConn{}
	ReceivedBuffer = make(chan message.Message)
	vnodeNum := 2
	var yamlConfig YamlConfig
	err := yamlConfig.ParseYaml("nodes.yml")
	checkError(err)
	for _, each := range yamlConfig.Servers {
		/* Build the NodeIndexMap hashmap 
		if each.Hostname = localHost {
			localHostIndex = index
		} */
		NodeIndexMap[each.Hostname] = node.NewNode(each.Hostname, each.Ip, each.Port, vnodeNum)
	}

	for key, value := range NodeIndexMap {
	    fmt.Println("Key:", key, "Value:", value)
	}

	LocalNode = NodeIndexMap[localHost]
	//LocalNode = yamlConfig.Servers[localHostIndex]
	fmt.Printf("Local Host is [%s] at %s:%d\n", LocalNode.Hostname, LocalNode.Ip, LocalNode.Port)
	fmt.Printf("Keys are: %+v\n", LocalNode.Keys)
	acceptConnectionsFromOtherServers(&yamlConfig)
	connectToOtherServers(&yamlConfig)
	fmt.Println("***************************************************")
	fmt.Println("****** Done all conections and back to main *******")
	fmt.Println("***************************************************")
}

func GetLocalNode() *node.Node {
	return LocalNode
}

func CheckIncomingMessages() message.Message {
	newMessage := <- ReceivedBuffer
	return newMessage
}

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
    }
}
