package swim

import (
	//"fmt"
	"github.com/otnt/ds/infra"
	"time"
	"github.com/otnt/ds/message"
	"log"
)

// Default configuration parameters
const (
	WAIT_TIME_DEFAULT     = 1000
	PING_INTERVAL_DEFAULT = 2000
	RAMDOM_PING_NUM = 3
	TASK_QUEUE_SIZE = 10
)

type task struct {
	req *message.Message
	res chan *message.Message
}

type SwimProtocol struct {
	// failure detector
	failureDetector *FailureDetector

	//PingChan chan *message.Message
	//PingAckChan chan *message.Message
	//ForwardChan chan *message.Message
	//ForwardAckChan chan *message.Message

	PingChan chan *message.Message
	AckChan chan *message.Message
	ForwardChan chan *message.Message

	taskChan chan *task
}


// Ping nodes at interval of PING_INTERVAL_DEFAULT
func (swim *SwimProtocol) Run() {
	swim.runTaskDoer()
	swim.runPinger()
	swim.runListener()
}

func (swim *SwimProtocol) runTaskDoer() {
	swim.taskChan = make(chan *task, TASK_QUEUE_SIZE)
	go func() {
		for {
			t := <-swim.taskChan
			infra.SendUnicast(t.req.Dest, t.req.Data, t.req.Kind)
			select {
			case rcv := <-swim.AckChan:
				t.res <- rcv
			case <-time.After(time.Millisecond * WAIT_TIME_DEFAULT):
				close(t.res)
			}
		}
	}()
}

// Ping next node at interval time, if direct ping fail, ask other nodes
// to ping it, if it fails again, esclate the suspect level of this node
func (swim *SwimProtocol) runPinger() {
	go func() {
		for {
			intervalChan := time.After(time.Millisecond * PING_INTERVAL_DEFAULT)
			if !swim.pingNext() {
				//if !swim.pingRandom() {
				//	swim.failureDetector.fail()
				//}
				swim.failureDetector.fail()
				log.Printf("failed %s\n", swim.failureDetector.curr)
			} else {
				log.Printf("succeed %s\n", swim.failureDetector.curr)
			}
			<-intervalChan
		}
	}()
}

// Swim protocol listener
func (swim *SwimProtocol) runListener() {
	swim.runPingResponse()
	//swim.runForwardResponse()
}

// Listen to direct ping request
func (swim *SwimProtocol) runPingResponse() {
	go func() {
		for {
			msg := <-swim.PingChan
			infra.SendUnicast(swim.failureDetector.ackMessage(msg))
			//log.Printf("Get ping from %s\n", msg.Src)
			swim.failureDetector.update(msg)
		}
	}()
}

//// Listen to ping forward request
//func (swim *SwimProtocol) runForwardResponse() {
//	go func() {
//		for {
//			//msg := <-swim.ForwardChan
//			//infra.SendUnicast(swim.failureDetector.forwardMessage(msg))
//			//select {
//			//case msg = <-swim.ForwardAckChan:
//			//	infra.SendUnicast(swim.failureDetector.ackMessage(msg))
//			//case <-time.After(time.Millisecond * WAIT_TIME_DEFAULT):
//			//}
//			msg := <-swim.ForwardChan
//
//			req := swim.failureDetector.forwardMessage(msg)
//			res := make(chan *message.Message)
//			swim.taskChan <- &task{req:req, res:res}
//
//			if rcv := <-res; rcv != nil {
//				swim.failureDetector.update(msg)
//				infra.SendUnicast(swim.failureDetector.ackMessage(msg))
//			}
//		}
//	}()
//}

// Ping next node, return true if ping succeed, false if fail
func (swim *SwimProtocol) pingNext() bool {
	//infra.SendUnicast(swim.failureDetector.nextMessage())

	//select {
	//case rcvMsg := <-swim.PingAckChan:
	//	swim.failureDetector.update(rcvMsg)
	//	return true
	//case <-time.After(time.Millisecond * WAIT_TIME_DEFAULT):
	//	return false
	//}

	req := swim.failureDetector.nextMessage()
	res := make(chan *message.Message)
	swim.taskChan <- &task{req:req, res:res}

	//log.Printf("%s ping %s\n", infra.LocalNode.Hostname, req.Dest)
	if rcv := <-res; rcv != nil {
		swim.failureDetector.update(rcv)
		return true
	} else {
		return false
	}
}

// Ping random several nodes
//func (swim *SwimProtocol) pingRandom() bool {
//	for i := 0; i<RAMDOM_PING_NUM; i++ {
//		go func() {
//			infra.SendUnicast(swim.failureDetector.randomMessage())
//		}()
//	}
//
//	select {
//	case rcvMsg := <-swim.PingAckChan:
//		swim.failureDetector.update(rcvMsg)
//		return true
//	case <-time.After(time.Millisecond * WAIT_TIME_DEFAULT):
//		return false
//	}
//}

// Create new swim protocol detector
func NewSwimProtocol(failureDetector *FailureDetector) *SwimProtocol {
	swim := &SwimProtocol{failureDetector:failureDetector}

	//swim.PingAckChan = make(chan *message.Message)
	swim.AckChan = make(chan *message.Message)
	swim.PingChan = make(chan *message.Message)
	swim.ForwardChan = make(chan *message.Message)
	return swim
}
