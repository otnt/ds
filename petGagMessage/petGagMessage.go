package petGagMessage

import (
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/petgagData"
)

type PetGagMessage struct {
	PGMessage *message.Message
	PGData    *petgagData.PetgagData
}

func NewPetGagMessage(msg *message.Message, data *petgagData.PetgagData) {
	newPetgagMessage := &PetGagMessage{}
	newPetgagMessage.PGMessage = msg
	newPetgagMessage.PGData = data
}
