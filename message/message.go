package message

import "encoding/json"

type Message struct {
	Src string
	Dest string
	Data string
	Kind string
}

func NewMessage(src string, dest string, data string, kind string) (message Message) {
	message = Message{}
	message.Src = src
	message.Dest = dest
	message.Data = data
	message.Kind = kind
	return message
}

func GetSrc(message *Message) string {
	return message.Src
}

func SetSrc(message *Message, src string) {
	message.Src = src
}

func GetDst(message *Message) string {
	return message.Dest
}

func SetDst(message *Message, dest string) {
	message.Dest = dest
}

func GetData(message *Message) string {
	return message.Data
}

func SetData(message *Message, data string) {
	message.Data = data
}

func GetKind(message *Message) string {
	return message.Kind
}

func SetKind(message *Message, kind string) {
	message.Kind = kind
}

func Marshal (message *Message) []byte {
	b, _ := json.Marshal(*message)
	return b
}

func Unmarshal (data []byte, message *Message) error {
	err := json.Unmarshal(data, message)
	return err
}
