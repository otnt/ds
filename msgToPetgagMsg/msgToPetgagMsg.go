package msgToPetgagMsg

import (
	"fmt"
	"github.com/pshastry/message"
	"github.com/pshastry/petGagMessage"
	"github.com/pshastry/petgagData"
	//"labix.org/v2/mgo/bson"
	"bytes"
	"encoding/json"
)

func ConvertToPGMsg(msg *message.Message) (pgData petGagMessage.PetGagMessage) {
	var decodedData petgagData.PetgagData
	err := json.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&decodedData)
	if err != nil {
		fmt.Printf("Error when decode data %v\n", err)
	}
	newPGmsg := petGagMessage.NewPetGagMessage(msg, &decodedData)
	return newPGmsg
}
