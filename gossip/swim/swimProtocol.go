package swim

import (
	"fmt"
	"github.com/otnt/ds/infra"
	"time"
	"github.com/otnt/ds/message"
	//"log"
	"github.com/otnt/ds/utils"
)

// Default configuration parameters
const (
	WAIT_TIME_DEFAULT     = 200
	PING_INTERVAL_DEFAULT = 400
	RAMDOM_PING_NUM = 2
	TASK_QUEUE_SIZE = 10
)

type task struct {
	req *message.Message
	res chan *message.Message
}

type SwimProtocol struct {
	// failure detector
	failureDetector *FailureDetector

	PingChan chan *message.Message
	AckChan chan *message.Message
	ForwardChan chan *message.Message
	ForwardAckChan chan *message.Message

	taskChan chan *task
	forwardTaskChan chan *task
}


// Ping nodes at interval of PING_INTERVAL_DEFAULT
func (swim *SwimProtocol) Run() {
	swim.runTaskDoer()
	swim.runForwardTaskDoer()

	time.Sleep(time.Millisecond * 2000)
	swim.runPinger()
	swim.runListener()
}

//func (swim *SwimProtocol) runTaskDoer() {
//	swim.taskChan = make(chan *task, TASK_QUEUE_SIZE)
//	mapper := make(map[string](chan (chan*message.Message)))
//
//	// put ack message to corresponding chan
//	go func() {
//		for {
//			rcv := <-swim.AckChan
//			src := rcv.Src
//			c := <-mapper[src]
//			c <- rcv
//		}
//	}()
//
//	// put task into channel and count and close channel
//	go func() {
//		for {
//			t := <-swim.taskChan
//			fmt.Println("get task")
//			c := mapper[t.req.Dest]
//			if c == nil {
//				c = make(chan (chan *message.Message), 1)
//				mapper[t.req.Dest] = c
//			}
//
//			go func() {
//				fmt.Println("chan chan ", c)
//				c <- t.res
//				infra.SendUnicast(t.req.Dest, t.req.Data, t.req.Kind)
//				fmt.Println("sent out")
//				<-time.After(time.Millisecond * WAIT_TIME_DEFAULT)
//				t.res <- nil
//			}()
//		}
//	}()
//}
//
//func (swim *SwimProtocol) runForwardTaskDoer() {
//	swim.forwardTaskChan= make(chan *task, TASK_QUEUE_SIZE)
//	mapper := make(map[string](chan (chan*message.Message)))
//
//	// put ack message to corresponding chan
//	go func() {
//		for {
//			rcv := <-swim.ForwardAckChan
//			src := rcv.Src
//			c := <-mapper[src]
//			c <- rcv
//		}
//	}()
//
//	// put task into channel and count and close channel
//	go func() {
//		for {
//			t := <-swim.forwardTaskChan
//			c := mapper[t.req.Dest]
//			if c == nil {
//				c = make(chan (chan *message.Message))
//				mapper[t.req.Dest] = c
//			}
//
//			go func() {
//				c <- t.res
//				infra.SendUnicast(t.req.Dest, t.req.Data, t.req.Kind)
//				<-time.After(time.Millisecond * WAIT_TIME_DEFAULT)
//				close(t.res)
//			}()
//		}
//	}()
//}

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
func (swim *SwimProtocol) runForwardTaskDoer() {
	swim.forwardTaskChan = make(chan *task, TASK_QUEUE_SIZE)
	go func() {
		for {
			t := <-swim.forwardTaskChan
			infra.SendUnicast(t.req.Dest, t.req.Data, t.req.Kind)
			select {
			case rcv := <-swim.ForwardAckChan:
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
			if !swim.pingNext() && !swim.pingRandom() {
					swim.failureDetector.fail(swim.failureDetector.curr)
			}
			<-intervalChan
		}
	}()
}

// Swim protocol listener
func (swim *SwimProtocol) runListener() {
	swim.runPingResponse()
	swim.runForwardResponse()
}

// Listen to direct ping request
func (swim *SwimProtocol) runPingResponse() {
	go func() {
		for {
			msg := <-swim.PingChan
			infra.SendUnicast(swim.failureDetector.ackMessage(msg))
			//fmt.Printf("Get ping from %s\n", msg.Src)
			swim.failureDetector.update(msg)
		}
	}()
}

// Listen to ping forward request
func (swim *SwimProtocol) runForwardResponse() {
	go func() {
		for {
			msg := <-swim.ForwardChan

			req := swim.failureDetector.forwardMessage(msg)
			res := make(chan *message.Message)
			swim.taskChan <- &task{req:req, res:res}

			//fmt.Printf("Get forward from %s to %s\n", msg.Src, msg.Dest)

			if rcv := <-res; rcv != nil {
				swim.failureDetector.update(msg)
				//infra.SendUnicast(swim.failureDetector.ackMessage(msg))
				infra.SendUnicast(swim.failureDetector.forwardAckMessage(msg))
			}
		}
	}()
}

// Ping next node, return true if ping succeed, false if fail
func (swim *SwimProtocol) pingNext() bool {
	req := swim.failureDetector.nextMessage()
	res := make(chan *message.Message)
	//fmt.Println("ping next task ", req)
	swim.taskChan <- &task{req:req, res:res}

	//log.Printf("%s ping %s\n", infra.LocalNode.Hostname, req.Dest)
	if rcv := <-res; rcv != nil {
		swim.failureDetector.update(rcv)
		fmt.Printf("ping next node %s succeed\n", rcv.Src)
		return true
	} else {
		fmt.Printf("ping next node %s failed\n",rcv.Src)
		return false
	}
}

// Ping random several nodes
func (swim *SwimProtocol) pingRandom() bool {
	fmt.Print("ping random nodes [ ")
	fo := make(chan *message.Message, RAMDOM_PING_NUM)
	for i := 0; i<RAMDOM_PING_NUM; i++ {
		req := swim.failureDetector.randomMessage()
		fmt.Print(req.Dest + " ")
		res := make(chan *message.Message)
		utils.Fanout(fo, res)
		//swim.forwardTaskChan <- &task{req:req, res:res}
	}
	fmt.Println("]")

	select {
	case rcvMsg := <-fo:
		if rcvMsg != nil {
			swim.failureDetector.update(rcvMsg)
			return true
		} else {
			return false
		}
	case <-time.After(time.Millisecond * WAIT_TIME_DEFAULT * RAMDOM_PING_NUM / 2):
		//fmt.Println("ping random nodes failed")
		return false
	}
}

// Create new swim protocol detector
func NewSwimProtocol(failureDetector *FailureDetector) *SwimProtocol {
	swim := &SwimProtocol{failureDetector:failureDetector}

	swim.AckChan = make(chan *message.Message)
	swim.PingChan = make(chan *message.Message)
	swim.ForwardChan = make(chan *message.Message)
	swim.ForwardAckChan = make(chan *message.Message)
	return swim
}
