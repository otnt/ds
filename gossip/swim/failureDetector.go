package swim

import (
	"sync"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/message"
	"encoding/json"
	"bytes"
	"log"
	"github.com/otnt/ds/infra"
	"fmt"
	"github.com/otnt/ds/replication"
	"math/rand"
	"time"
)

const (
	FAIL_TIME = 5000
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
	Forward string
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
	if fd.index == 0 {
		shuffle(fd.hostnames)
	}
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
func (fd *FailureDetector) randomMessage() (*message.Message) {
	randNode := fd.hostnames[rand.Intn(len(fd.hostnames))]
	for randNode == infra.LocalNode.Hostname || randNode == fd.curr || fd.info[randNode] == FAULTY {
		randNode = fd.hostnames[rand.Intn(len(fd.hostnames))]
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(
		failureDetectorMessage{
			Src:infra.LocalNode.Hostname,
			Forward:randNode,
			Info:fd.info,
		},
	)

	//return randNode, buf.String(), SWIM_FORWARD
	return &message.Message{Dest:randNode, Data:buf.String(), Kind:SWIM_FORWARD}
}

func (fd *FailureDetector) ackMessage(msg *message.Message) (string, string, string) {
	fdm := failureDetectorMessage{
		Src: infra.LocalNode.Hostname,
		Info:fd.info,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(fdm)
	return msg.Src, buf.String(), SWIM_ACK
}

func (fd *FailureDetector) forwardAckMessage(msg *message.Message) (string, string, string) {
	fdm := failureDetectorMessage{
		Src: infra.LocalNode.Hostname,
		//forward: "",
		Info:fd.info,
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(fdm)
	return msg.Src, buf.String(), SWIM_FORWARD_ACK
}

//func (fd *FailureDetector) forwardMessage(msg *message.Message) (string, string, string) {
func (fd *FailureDetector) forwardMessage(msg *message.Message) *message.Message {
	var info failureDetectorMessage
	json.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&info)

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(
		failureDetectorMessage{
			Src:infra.LocalNode.Hostname,
			Info:fd.info,
		},
	)

	//return info.pingNode.Hostname, buf.String(), SWIM_PING
	return &message.Message{Dest:info.Forward, Kind:SWIM_PING, Data:buf.String()}
}

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

	//fd.updateStatus(fdm.Src, HEALTHY)
	fd.updateStatus(fdm.Src, RECOVER)
}

func (fd *FailureDetector) updateStatus(hostname string, status string) {
	//fd.info[hostname] = status
	if status == FAULTY {
		fd.info[hostname] = FAULTY
	} else if status == SUSPECTED && fd.info[hostname] == HEALTHY {
		fd.info[hostname] = SUSPECTED
	} else if status == RECOVER {
		fd.info[hostname] = HEALTHY
	}
}

const (
	HEALTHY = "healthy"
	//SUSPECTED_1 = "suspected_1"
	//SUSPECTED_2 = "suspected_2"
	//SUSPECTED_3 = "suspected_3"
	RECOVER = "recover"
	SUSPECTED = "suspected"
	FAULTY = "faulty"
)

//var statusFailMap = map[string]string{
//	HEALTHY:SUSPECTED_1,
//	SUSPECTED_1:SUSPECTED_2,
//	SUSPECTED_2:SUSPECTED_3,
//	SUSPECTED_3:FAULTY,
//	FAULTY:FAULTY,
//}

// Current node probe failed
//func (fd *FailureDetector) fail() {
//	newStatus := statusFailMap[fd.info[fd.curr]]
//	fd.info[fd.curr] = newStatus
//
//	if newStatus == FAULTY {
//		fmt.Printf("node %s failed, now replicating...\n", fd.curr)
//
//		replication.NotifyNodeDies(infra.NodeIndexMap[fd.curr].Keys[0])
//
//		fd.ring.RemoveSync(infra.NodeIndexMap[fd.curr])
//	}
//	log.Printf("%s new status %s\n", fd.curr, newStatus)
//}

var mutex sync.Mutex
func (fd *FailureDetector) fail(name string) {
	go func() {
		fd.info[name] = SUSPECTED
		<-time.After(time.Millisecond * FAIL_TIME)
		newStatus := fd.info[name]
		if fd.info[name] != HEALTHY {
			newStatus = FAULTY
			fd.info[name] = FAULTY
		}

		if newStatus == FAULTY {
			mutex.Lock()
			_, found := fd.ring.Get(infra.NodeIndexMap[name].Keys[0]) 
			if found {
				fmt.Printf("******node %s failed, now replicating...********\n", name)

				replication.NotifyNodeDies(infra.NodeIndexMap[name].Keys[0])

				fd.ring.RemoveSync(infra.NodeIndexMap[name])
			}
			mutex.Unlock()
		}
	}()
}

func shuffle(list []string) {
	for i := 1; i < len(list); i++ {
		ii := rand.Intn(i)
		tmp := list[ii]
		list[ii] = list[i]
		list[i] = tmp
	}
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
	shuffle(fd.hostnames)

	return &fd
}
