package utils

import (
	"github.com/otnt/ds/message"
)

func Fanout(out chan *message.Message, in chan *message.Message) {
	go func() {
		msg := <-in
		out<-msg
	}()
}
