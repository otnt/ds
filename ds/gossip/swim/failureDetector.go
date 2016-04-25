package swimProtocol

import (
	"github.com/otnt/ds/node"
	ch "github.com/otnt/ds/consistentHashing"
	"math/rand"
	"github.com/otnt/ds/message"
	"encoding/gob"
	"bytes"
	"log"
	"github.com/otnt/ds/infra"
)

const (
	SWIM_PING = "swim_ping"
	SWIM_RANDOM = "swim_random"
	SWIM_ACK = "swim_ack"
)

// Send this message to other failure detector
type failureDetectorMessage struct {
	localNode *node.Node
	pingNode *node.Node
	information []*node.Node
}

type FailureDetector struct {
	index int
	curr *node.Node
	nodes []*node.Node
	ring *ch.Ring
}

// Get next ping host, and ping data
func (fd *FailureDetector) nextMessage() (string, string, string) {
	if fd.index == len(fd.nodes) {
		fd.refresh()
	}

	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(
		failureDetectorMessage{
			localNode:infra.LocalNode,
			pingNode:nil,
			information:fd.nodes,
		},
	)

	fd.curr = fd.nodes[fd.index]
	return fd.curr.Hostname, SWIM_PING, buf.String()
}

// Get random ping host, and ping data
func (fd *FailureDetector) randomMessage() (string, string, string) {
	randNode := fd.nodes[rand.Intn(len(fd.nodes))]

	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(
		failureDetectorMessage{
			localNode:infra.LocalNode,
			pingNode:randNode,
			information:fd.nodes,
		},
	)

	return randNode.Hostname, SWIM_RANDOM, buf.String()
}

func (fd *FailureDetector) ackMessage(msg *message.Message) (string, string, string) {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(
		failureDetectorMessage{
			localNode:infra.LocalNode,
			pingNode:nil,
			information:fd.nodes,
		},
	)
	return msg.Src, SWIM_ACK, buf.String()
}

func (fd *FailureDetector) forwardMessage(msg *message.Message) (string, string, string) {
	var info failureDetectorMessage
	_= gob.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&info)

	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(
		failureDetectorMessage{
			localNode:infra.LocalNode,
			pingNode:nil,
			information:fd.nodes,
		},
	)

	return info.pingNode.Hostname, SWIM_PING, buf.String()
}

// Update node status
func (fd *FailureDetector) update(msg *message.Message) {
	var info failureDetectorMessage
	err := gob.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&info)
	if err != nil {
		log.Printf("Error when updating %+v\n", err)
		return
	}

	for _, n := range info.information{
		fd.ring.UpdateStatus(n)
	}

	fd.ring.UpdateStatus(info.localNode)
}

const (
	HEALTHY = "healthy"
	SUSPECTED_1 = "suspected_1"
	SUSPECTED_2 = "suspected_2"
	SUSPECTED_3 = "suspected_3"
	FAULTY = "faulty"
)

var statusFailMap = map[string]string{
	HEALTHY:SUSPECTED_1,
	SUSPECTED_1:SUSPECTED_2,
	SUSPECTED_2:SUSPECTED_3,
	SUSPECTED_3:FAULTY,
}

// Current node probe failed
func (fd *FailureDetector) fail() {
	fd.curr.Status = statusFailMap[fd.curr.Status.(string)]
	fd.ring.UpdateStatus(fd.curr)
	if fd.curr.Status.(string) == FAULTY {
		fd.ring.RemoveSync(fd.curr)
	}
}

// Refresh node list
func (fd *FailureDetector) refresh() {
	fd.nodes = fd.ring.Values()
}



func NewFailureDetector(ring *ch.Ring) *FailureDetector {
	fd := FailureDetector{index:0, ring:ring}
	fd.refresh()
	return &fd
}
