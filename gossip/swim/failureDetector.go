package swim

import (
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/message"
	"encoding/json"
	"bytes"
	"log"
	"github.com/otnt/ds/infra"
	"fmt"
	"github.com/otnt/ds/replication"
)

const (
	SWIM_PING = "swim_ping"
	SWIM_FORWARD = "swim_forward"
	SWIM_FORWARD_ACK = "swim_forward_ack"
	SWIM_ACK = "swim_ack"
)

// Send this message to other failure detector
type failureDetectorMessage struct {
	Src string
	//forward string
	Info map[string]string
}

type FailureDetector struct {
	index int
	curr string
	ring *ch.Ring
	info map[string]string
	hostnames []string
}

// Get next ping host, and ping data
func (fd *FailureDetector) nextMessage() *message.Message {
	fdm := &failureDetectorMessage{
		Src: infra.LocalNode.Hostname,
		Info:fd.info,
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(fdm)
	if err != nil {
		panic(err)
	}

	fd.index = (fd.index + 1) % len(fd.hostnames)
	fd.curr = fd.hostnames[fd.index]
	for fd.curr == infra.LocalNode.Hostname || fd.info[fd.curr] == FAULTY {
		fd.index = (fd.index + 1) % len(fd.hostnames)
		fd.curr = fd.hostnames[fd.index]
	}

	msg := &message.Message{Dest:fd.curr, Data:buf.String(), Kind:SWIM_PING}
	return msg
}

//// Get random ping host, and ping data
//func (fd *FailureDetector) randomMessage() (string, string, string) {
//	randNode := fd.nodes[rand.Intn(len(fd.nodes))]
//
//	var buf bytes.Buffer
//	gob.NewEncoder(&buf).Encode(
//		failureDetectorMessage{
//			localNode:infra.LocalNode,
//			pingNode:randNode,
//			information:fd.nodes,
//		},
//	)
//
//	return randNode.Hostname, buf.String(), SWIM_FORWARD
//}

func (fd *FailureDetector) ackMessage(msg *message.Message) (string, string, string) {
	fdm := failureDetectorMessage{
		Src: infra.LocalNode.Hostname,
		//forward: "",
		Info:fd.info,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(fdm)
	return msg.Src, buf.String(), SWIM_ACK //SWIM_PING_ACK
}

////func (fd *FailureDetector) forwardMessage(msg *message.Message) (string, string, string) {
//func (fd *FailureDetector) forwardMessage(msg *message.Message) *message.Message {
//	var info failureDetectorMessage
//	_= gob.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&info)
//
//	var buf bytes.Buffer
//	gob.NewEncoder(&buf).Encode(
//		failureDetectorMessage{
//			localNode:infra.LocalNode,
//			pingNode:nil,
//			information:fd.nodes,
//		},
//	)
//
//	//return info.pingNode.Hostname, buf.String(), SWIM_PING
//	return &message.Message{Dest:info.pingNode.Hostname, Kind:SWIM_PING, Data:buf.String()}
//}

// Update node status
func (fd *FailureDetector) update(msg *message.Message) {
	var fdm failureDetectorMessage
	err := json.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&fdm)
	//log.Printf("info is %+v\n",fdm)
	if err != nil {
		log.Printf("Error when updating %+v\n", err)
		return
	}

	for k, v := range fdm.Info{
		fd.updateStatus(k, v)
	}

	fd.updateStatus(fdm.Src, HEALTHY)
}

func (fd *FailureDetector) updateStatus(hostname string, status string) {
	fd.info[hostname] = status
}

const (
	HEALTHY = "healthy"
	SUSPECTED_1 = "suspected_1"
	SUSPECTED_2 = "suspected_2"
	SUSPECTED_3 = "suspected_3"
	FAULTY = "faulty"
)

const (
	INIT_STATUS = HEALTHY
)

var statusFailMap = map[string]string{
	HEALTHY:SUSPECTED_1,
	SUSPECTED_1:SUSPECTED_2,
	SUSPECTED_2:SUSPECTED_3,
	SUSPECTED_3:FAULTY,
	FAULTY:FAULTY,
}

// Current node probe failed
func (fd *FailureDetector) fail() {
	newStatus := statusFailMap[fd.info[fd.curr]]
	fd.info[fd.curr] = newStatus

	if newStatus == FAULTY {
		fmt.Printf("node %s is failed, now replicating...\n", fd.curr)

		replication.NotifyNodeDies(infra.NodeIndexMap[fd.curr].Keys[0])

		fd.ring.RemoveSync(infra.NodeIndexMap[fd.curr])
	}
	log.Printf("%s new status %s\n", fd.curr, newStatus)
}

func NewFailureDetector(ring *ch.Ring) *FailureDetector {
	fd := FailureDetector{index:0, ring:ring, curr:infra.LocalNode.Hostname}
	nodes := ring.Values()
	fd.hostnames = make([]string, 0)
	fd.info = make(map[string]string)

	for _, n := range nodes {
		fd.hostnames = append(fd.hostnames, n.Hostname)
		fd.info[n.Hostname] = n.Status.(string)
	}
	fd.index = 0

	return &fd
}
