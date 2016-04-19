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


import (
	"fmt"
	"io/ioutil"
	"os"
	"net"
	"strings"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/node"
	"github.com/otnt/ds/mongoDBintegration"
	"github.com/otnt/ds/PetGagData"
	"gopkg.in/yaml.v2"
	"strings"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/ReceiverThread"
)

type replicaNode struct {
	const replicationFactor int = 2
	numAcks int
}

/* Getter and Setter for numAcks */


func initMongoDB() (mongoSession *mgo.Session) {
	mongoSession = establishSession()
}

/* Functions to be implemented if the message is of kind forward */

func updateSelfDB(message PetGagMessage) (objIDHex string){
	switch message.PetGagData.DbOp {
		case Insert:
			objID := insertPicture(mongoSession, message.PetGagData.ImageURL, message.PetGagData.UserName, message.PetGagData.BelongsTo)
			message.ObjID = objID
			return objID.Hex

		case Upvote:
			err := upVotePicture(mongoSession, ObjectIdHex(message.PetGagData.ObjID), message.PetGagData.UserName, message.PetGagData.UpVote, message.PetGagData.BelongsTo)
			if (err != nil) {
				fmt.Println("Error in upvoting picture, %s\n", err)
			}
			break;

		case Downvote:
			err := downVotePicture(mongoSession, ObjectIdHex(message.PetGagData.ObjID), message.PetGagData.UserName, message.PetGagData.DownVote, message.PetGagData.BelongsTo)
			if (err != nil) {
				fmt.Println("Error in downvoting picture, %s\n", err)
			}
			break;

		case Comment:
			err := commentOnPicture(mongoSession, ObjectIdHex(message.PetGagData.ObjID), message.PetGagData.UName, message.PetGagData.Comment, message.PetGagData.BelongsTo)
			if (err != nil) {
				fmt.Println("Error in commenting on picture, %s\n", err)
			}
			break;

		case Delete:
			err := deleteFromDB(mongoSession , ObjectIdHex(message.PetGagData.ObjID), message.PetGagData.BelongsTo)
			if (err != nil) {
				fmt.Println("Error in deleting picture, %s\n", err)
			}
			break;

		default:
			fmt.Println("Enter the correct operation: Insert / Upvote / Downvote / Comment / Delete")
			break;
	}
}

func AskNodesToUpdate(message PetGagMessage) {
	var secondaryNodes []node.Node
	var localNode node.Node = GetLocalNode()
	secondaryNodes = append(secondaryNodes, localNode)

	for(int i = 1; i <= replicationFactor; i++) {
		newNode, err := Successor(secondaryNodes[i-1])
		if (err != nil) {
			fmt.Println("Error in obtaining successor node %s",err)
		}
		secondaryNodes = append(secondaryNodes, newNode)
	}

	for(int i = 1; i < replicationFactor; i++) {
		SendUnicast(secondaryNodes[i], message.Data, "replication")
	}

}

func WaitForAcks() {
	var int acksObtained = 0
	for() {
		acksObtained = 

	}
}

func processAcks(message PetGagMessage) {


}

func respondToClient() {


}

/* Functions to be implemented if the message is of kind replicate */
func sendAcks(message PetGagMessage) {
	SendUnicast(message.Src, "Received", "acknowledgement")

}

/* Reuse updateSelfDB() */


