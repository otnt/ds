/** Replication for write operations **/
/* Methods required */
/* Check if the node is a primary node */
/* If it is not primary ==> inform the primary of a new request from client ==> Skip this step if it is primary */
/* Primary updates its own database and informs all the other nodes in its network to update */
/* Once a node completes the update, it sends its acknowledgements to the primary */
/* The primary then informs the requesting node that all databases have been updated */
/* The node then serves the client */

/* Questions */
/* Role of loadbalancer in this system */
/* What happens when one of the nodes fails and doesnt send its ack? */
/* Should it contact the loadbalancer every time it needs to know the list of nodes */

/********************************************************************************************************************/

/* Add a field original_source to the data => signifies which node the data should belong to */
package replication

import (
	"fmt"
	ch "github.com/pshastry/consistentHashing"
	"github.com/pshastry/infra"
	"github.com/pshastry/mongoDBintegration"
	//"github.com/otnt/ds/node"
	"github.com/pshastry/petGagMessage"
	//"github.com/otnt/ds/petgagData"
	"bytes"
	"encoding/json"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type DbOperation int

const (
	Insert DbOperation = 1 + iota
	Upvote
	Downvote
	Comment
	Delete
)

const replicationFactor int = 2

var numAcks int = 0
var ring *ch.Ring

func InitReplication(r *ch.Ring) {
	ring = r
}

/* func InitMongoDB() (mongoSession *mgo.Session) {
	mongoSession = mongoDBintegration.EstablishSession()
	return mongoSession
} */

/* Functions to be implemented if the message is of kind forward */

func UpdateSelfDB(msg *petGagMessage.PetGagMessage, mongoSession *mgo.Session) (objIDHex string) {
	operation := msg.PGData.DbOp

	if operation == "Insert" {

		/* Each user has his own collection. Collections are names after users */
		objID := mongoDBintegration.InsertPicture(mongoSession, msg.PGData.ImageURL, msg.PGData.UserName, msg.PGData.BelongsTo, msg.PGData.ObjID)
		msg.PGData.ObjID = objID.Hex()
		return msg.PGData.ObjID
	}

	if operation == "Upvote" {
		mongoDBintegration.UpVotePicture(mongoSession, bson.ObjectIdHex(msg.PGData.ObjID), msg.PGData.UpVote, msg.PGData.BelongsTo)
		return "success"
	}

	if operation == "Downvote" {
		mongoDBintegration.DownVotePicture(mongoSession, bson.ObjectIdHex(msg.PGData.ObjID), msg.PGData.DownVote, msg.PGData.BelongsTo)
		return "success"
	}

	if operation == "Comment" {
		comment := msg.PGData.Commt
		mongoDBintegration.CommentOnPicture(mongoSession, bson.ObjectIdHex(msg.PGData.ObjID), comment[len(comment)-1].UserName, comment[len(comment)-1].Comt, msg.PGData.BelongsTo)
		return "success"
	}

	if operation == "Delete" {
		mongoDBintegration.DeleteFromDB(mongoSession, bson.ObjectIdHex(msg.PGData.ObjID), msg.PGData.BelongsTo)
		return "success"
	} else {
		fmt.Println("Enter the correct operation: Insert / Upvote / Downvote / Comment / Delete")
		return ""
	}
}

func AskNodesToUpdate(message *petGagMessage.PetGagMessage, mongoSession *mgo.Session) {
	//var secondaryNode node.Node
	var secNodeKeys []string
	//var localNode *node.Node = infra.GetLocalNode()
	var nodeKey string = GetKey(message)

	secNodeKeys = append(secNodeKeys, nodeKey)

	for i := 1; i <= replicationFactor; i++ {
		_, newNodeKey, err := ring.Successor(secNodeKeys[i-1])
		if err != nil {
			fmt.Println("Error in obtaining successor node %s", err)
		}
		secNodeKeys = append(secNodeKeys, newNodeKey)
	}
	for i := 1; i < replicationFactor; i++ {
		secondaryNode, err := ring.Get(secNodeKeys[i])
		if err == false {
			fmt.Println("Error in obtaining node from key: %s\n", err)
		}
		encoded_data := StructToString(message)
		infra.SendUnicast(secondaryNode.Hostname, encoded_data, "replication")
	}

}

func GetKey(message *petGagMessage.PetGagMessage) string {
	data := StructToString(message)
	dataKey := ring.Hash(data)
	_, currentKey, err := ring.LookUp(dataKey)
	if err != nil {
		fmt.Println("Error in obtaining key")
	}
	return currentKey
}

func WaitForAcks() {
	var acksObtained int = 0
	for {
		acksObtained = numAcks
		if acksObtained == replicationFactor {
			break
		}
	}
}

func ProcessAcks() {
	numAcks = numAcks + 1
}

func RespondToClient() {
	//infra.SendUnicast(message.PGData.BelongsTo, "Completed Replication", "response")
	fmt.Println("Replication is now complete")
}

/* Functions to be implemented if the message is of kind replicate */
func SendAcks(message *petGagMessage.PetGagMessage) {
	infra.SendUnicast(message.PGMessage.Src, "Received", "acknowledgement")

}

/***********************************************************************/

func StructToString(message *petGagMessage.PetGagMessage) (encoded_msg string) {
	var buf bytes.Buffer
	data := message.PGData
	json.NewEncoder(&buf).Encode(data)
	return buf.String()
}
