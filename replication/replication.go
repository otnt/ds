/** Replication for write operations **/
/* Methods required */
/* Check if the node is a primary node */
/* If it is not primary ==> inform the primary of a new request from client ==> Skip this step if it is primary */
/* Primary updates its own database and informs all the other nodes in its network to update */
/* Once a node completes the update, it sends its acknowledgements to the primary */
/* The primary then informs the requesting node that all databases have been updated */
/* The node then serves the client */

package replication

import (
	"bytes"
	"encoding/json"
	"fmt"
	ch "github.com/otnt/ds/consistentHashing"
	"github.com/otnt/ds/dbAccess"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/message"
	"log"
	//"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
)

type DbOperation int

const (
	INSERT           = "Insert"
	UPVOTE           = "Upvote"
	DOWNVOTE         = "Downvote"
	COMMENT          = "Comment"
	DELETE           = "Delete"
	KIND_REPLICATION = "Replication"
	KIND_REPLICN_ACK = "Replcn_Ack"
)

const ReplicationFactor int = 1

var NumAcks int = 0
var ReplRing *ch.Ring
var ReplChan chan *message.Message
var AckChan chan *message.Message

func InitReplication(r *ch.Ring) {
	ReplRing = r
	ReplChan = make(chan *message.Message)
	AckChan = make(chan *message.Message)
	initListener()
}

func initListener() {
	go func() {
		for {
			select {
			case msg := <-ReplChan:
				fmt.Println("Calling UpdateSelfDB\n")
				err := UpdateSelfDB(msg)
				fmt.Println("Updating DB", err)
				if err == nil {
					fmt.Println("Sending Acknowledgement")
					SendAcks(msg)
				}
			case msg := <-AckChan:
				fmt.Println("Incrementing Acknowledgement")
				ProcessAcks(msg)
			}
		}
	}()
}

/* Receives the dbAccess.PetGagPost */
/* Modifies the dbAccess.PetGagPost to a message to send it to secondary nodes */
func AskNodesToUpdate(post dbAccess.PetGagPost) {

	NumAcks = 0

	var secNodeKeys []string
	var nodeKey string = GetKey(post.ImageURL)

	secNodeKeys = append(secNodeKeys, nodeKey)

	fmt.Println("Converting post to a message")
	data := StructToString(post)

	fmt.Println("Successfully converted to string")

	/* Find Secondary Keys */
	for i := 1; i <= ReplicationFactor; i++ {
		_, newNodeKey, err := ReplRing.Successor(secNodeKeys[i-1])
		if err != nil {
			fmt.Println("Error in obtaining successor node %s", err)
		}
		secNodeKeys = append(secNodeKeys, newNodeKey)
	}

	/* Find secondary nodes from the keys */
	for i := 1; i <= ReplicationFactor; i++ {
		secondaryNode, err := ReplRing.Get(secNodeKeys[i])
		if err == false {
			fmt.Println("Error in obtaining node from key: %s\n", err)
		}
		fmt.Println("Sending to secondary nodes", secondaryNode.Hostname)
		infra.SendUnicast(secondaryNode.Hostname, data, KIND_REPLICATION)
		fmt.Println("Successfully sent to secondary nodes")
	}

	// Wait For Acknowledgements - Include time out option
	/*go func() {
		var acksObtained int = 0
		fmt.Println("Waiting For Acks")
		for {
			acksObtained = NumAcks
			//fmt.Println("Acks Obtained = ", NumAcks)
			if acksObtained == ReplicationFactor {
				fmt.Println("acks obtained = ", NumAcks)
				break
			}
		}
		RespondToClient()
	}()*/

}

func ProcessAcks(msg *message.Message) {
	fmt.Println("Inside Process Acks")
	NumAcks = NumAcks + 1
}

func RespondToClient() {
	//infra.SendUnicast(message.PGData.BelongsTo, "Completed Replication", "response")
	fmt.Println("Replication is now complete\n")
	fmt.Println("*************************************************")
}

/* Functions to be implemented if the message is of kind replicate */

func UpdateSelfDB(msg *message.Message) (err error) {
	/* Decode the message to get the Post */
	fmt.Println("Converting from string to struct")
	post := StringToStruct(msg.Data)

	fmt.Println("Successfully decoded into struct", post)
	operation := post.DbOp

	/* Modify Collection Name */

	post.BelongsTo = msg.Src + "-replication"
	fmt.Println("I am the secondary", post.BelongsTo)

	if operation == INSERT {
		_, err := post.Write()
		fmt.Println("Inserted into Database at the secondary")
		if err != nil {
			fmt.Println("Unable to insert into secondary db")
			log.Fatal(err)
			return err
		}
	} else if operation == COMMENT {
		addCommentMsg := makeCommentMsg(post)
		err := addCommentMsg.AddCommmentInDB()
		if err != nil {
			log.Fatal(err)
			return err
		}

	} else if operation == UPVOTE {
		voteMsg := makeVoteMsg(post)
		fmt.Println("Collection is : ", post.BelongsTo)
		err := voteMsg.UpvotePost()
		if err != nil {
			log.Fatal(err)
			return err
		}

	} else if operation == DOWNVOTE {
		voteMsg := makeVoteMsg(post)
		err := voteMsg.DownvotePost()
		if err != nil {
			log.Fatal(err)
			return err
		}
	} else {
		fmt.Println("Error in operation field\n")
	}

	return nil

}

func SendAcks(message *message.Message) {
	infra.SendUnicast(message.Src, "DB UPDATE COMPLETED", KIND_REPLICN_ACK)

}

/******************** Helper Functions ********************/

func StructToString(post dbAccess.PetGagPost) (encoded_msg string) {
	//	fmt.Println("Converting from struct to string")
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(post)
	fmt.Println("The string is: ", buf.String())
	return buf.String()
}

func GetKey(id string) string {

	dataKey := ReplRing.Hash(id)
	_, currentKey, err := ReplRing.LookUp(dataKey)

	if err != nil {
		fmt.Println("Error in obtaining key")
	}
	return currentKey
}

func StringToStruct(data string) (post dbAccess.PetGagPost) {
	err := json.NewDecoder(bytes.NewBufferString(data)).Decode(&post)
	if err != nil {
		fmt.Printf("Error when decode data %v\n", err)
	}
	return post
}

/*func getCollectionName() (collName string) {
	localNode := infra.GetLocalNode()
	collName = localNode.Hostname + "-replication"
	return localNode.Hostname
}*/

func makeVoteMsg(post dbAccess.PetGagPost) (voteMsg dbAccess.VoteMsg) {

	fmt.Println("Constructing vote message")
	voteMsg = dbAccess.VoteMsg{}
	voteMsg.BelongsTo = post.BelongsTo
	voteMsg.DbOp = post.DbOp
	voteMsg.ImageId = post.ObjID
	voteMsg.ImageURL = post.ImageURL
	return voteMsg
}

func makeCommentMsg(post dbAccess.PetGagPost) (commentMsg dbAccess.AddCommentMsg) {
	commentMsg = dbAccess.AddCommentMsg{}
	commentMsg.BelongsTo = post.BelongsTo
	commentMsg.DbOp = post.DbOp
	commentMsg.ImageId = post.ObjID
	commentMsg.ImageURL = post.ImageURL
	commentMsg.UName = post.CommentList[0].UserCName
	commentMsg.Comment = post.CommentList[0].Comment
	return commentMsg
}
