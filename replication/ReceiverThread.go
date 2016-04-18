package ReceiverThread

import {
	"fmt"
	"infra"
	"os"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/Replication"
}

func receiveForwardMessageThread(message PetGagMessage) {
	for {
		newMessage := infra.CheckIncomingMessages()
		var kind string = message.GetKind(&newMessage)
		if (kind == "forward") {
			id := updateSelfDB(newMessage)
			askNodesToUpdate(newMessage)
			go waitForAcks()
			respondToClient()
		}
		else if (kind == "replicate") {
			id := updateSelfDB(newMessage)
			sendAcks(newMessage)
		}
		else if (kind == "acknowledge") {
			processAcks(newMessage)

		}
	}
}

