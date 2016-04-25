package swimProtocol

import (
	//"fmt"
	"github.com/otnt/ds/infra"
	"time"
	"github.com/otnt/ds/message"
)

// Default configuration parameters
const (
	WAIT_TIME_DEFAULT     = 1000
	PING_INTERVAL_DEFAULT = 2000
	RAMDOM_PING_NUM = 3
)

type SwimProtocol struct {
	// failure detector
	failureDetector *FailureDetector

	AckChan chan *message.Message
	PingChan chan *message.Message
	ForwardChan chan *message.Message
}


// Ping nodes at interval of PING_INTERVAL_DEFAULT
func (swim *SwimProtocol) Run() {
	swim.runPinger()
	swim.runListener()
}

// Ping next node at interval time, if direct ping fail, ask other nodes
// to ping it, if it fails again, esclate the suspect level of this node
func (swim *SwimProtocol) runPinger() {
	intervalChan := time.After(time.Millisecond * PING_INTERVAL_DEFAULT)
	go func() {
		if !swim.pingNext() {
			if !swim.pingRandom() {
				swim.failureDetector.fail()
			}
		}
		<-intervalChan
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
			swim.failureDetector.update(msg)
		}
	}()
}

// Listen to ping forward request
func (swim *SwimProtocol) runForwardResponse() {
	go func() {
		for {
			msg := <-swim.ForwardChan
			infra.SendUnicast(swim.failureDetector.forwardMessage(msg))
			swim.failureDetector.update(msg)
		}
	}()
}

// Ping next node, return true if ping succeed, false if fail
func (swim *SwimProtocol) pingNext() bool {
	infra.SendUnicast(swim.failureDetector.nextMessage())

	select {
	case rcvMsg := <-swim.AckChan:
		swim.failureDetector.update(rcvMsg)
		return true
	case <-time.After(time.Millisecond * WAIT_TIME_DEFAULT):
		return false
	}
}

// Ping random several nodes
func (swim *SwimProtocol) pingRandom() bool {
	for i := 0; i<RAMDOM_PING_NUM; i++ {
		go func() {
			infra.SendUnicast(swim.failureDetector.randomMessage())
		}()
	}

	select {
	case rcvMsg := <-swim.AckChan:
		swim.failureDetector.update(rcvMsg)
		return true
	case <-time.After(time.Millisecond * WAIT_TIME_DEFAULT):
		return false
	}
}

// Create new swim protocol detector
func NewSwimProtocol(failureDetector *FailureDetector) *SwimProtocol {
	swim := &SwimProtocol{failureDetector:failureDetector}
	return swim
}
