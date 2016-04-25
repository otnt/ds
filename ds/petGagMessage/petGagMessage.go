package petGagMessage

import (
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/petgagData"
)

type PetGagMessage struct {
	PGMessage *message.Message
	PGData    *petgagData.PetgagData
}

func NewPetGagMessage(msg *message.Message, data *petgagData.PetgagData) PetGagMessage {
	newPGMessage := PetGagMessage{}
	newPGMessage.PGMessage = msg
	newPGMessage.PGData = data
	return newPGMessage
}
