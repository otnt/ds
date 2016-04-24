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
	db "github.com/otnt/ds/dbAccess"
	"github.com/otnt/ds/infra"
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/node"
	"log"
	"reflect"
	//"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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
	KIND_TRANSFER    = "Replcn_Transfer"
)

const ReplicationFactor int = 1

var NumAcks int = 0
var ReplRing *ch.Ring
var ReplChan chan *message.Message
var AckChan chan *message.Message
var TransferChan chan *message.Message

func InitReplication(r *ch.Ring) {
	ReplRing = r
	ReplChan = make(chan *message.Message)
	AckChan = make(chan *message.Message)
	TransferChan = make(chan *message.Message)
	initListener()
}

func UpdateNumAcks() {
	msg := <-AckChan
	ProcessAcks(msg)
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
			case msg := <-TransferChan:
				fmt.Println("Received a transfer message")
				AddNewCollection(msg)
			}
		}
	}()
}

func AskNodesToUpdate(post db.PetGagPost) {

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

/******************* Fault Tolerance APIs *******************************/

func AmIPredecessor(diedKey string) bool { /* The argument passed refers to the key of the node that has died */
	predNode, _, _ := ReplRing.Predecessor(diedKey)
	/* predNode, found := ReplRing.Get(predKey)
	if !found {
		fmt.Println("Error in getting node from key - Predecessor not found")
	} */

	localNode := infra.GetLocalNode()

	if reflect.DeepEqual(predNode, localNode) {
		return true
	} else {
		return false
	}
}

func AmISuccessor(diedKey string) bool {
	succNode, _, _ := ReplRing.Successor(diedKey)

	localNode := infra.GetLocalNode()

	if reflect.DeepEqual(succNode, localNode) {
		return true
	} else {
		return false
	}
}

func FindNewPrimary(diedKey string) (pnode *node.Node, pkey string) {
	succNode, succKey, _ := ReplRing.Successor(diedKey)
	return succNode, succKey
}

func SendCollnToNewPrimary(pnode *node.Node) {
	collection_name := infra.GetLocalNode().Hostname
	posts := db.GetAllPostsFromDB(collection_name)
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(posts)

	fmt.Println("Posts obtained are: ", posts)
	fmt.Println("Data encoded is: ", buf.String())
	/* Kind is assigned KIND_TRANSFER because this data should not be re-replicated */
	infra.SendUnicast(pnode.Hostname, buf.String(), KIND_TRANSFER)
}

func AddNewCollection(msg *message.Message) {
	collection_name := msg.Src + "-replication"
	var d []bson.M

	fmt.Println("Inside add new collection")

	err := json.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&d)
	if err != nil {
		fmt.Printf("Error when decode data %v\n", err)
		return
	}

	fmt.Println("Decoded data. Adding to collection")
	for _, post := range d {
		err := db.TransferToDb(collection_name, post)
		if err != nil {
			fmt.Println("Error in transferring data")
		}
	}
}

func MergeCollections(diedKey string) {
	diedNode, found := ReplRing.Get(diedKey)
	if !found {
		fmt.Println("died key not found")
	}
	collection_name := diedNode.Hostname + "-replication"
	fmt.Println("Inside Merge Collections")
	posts := db.GetAllPostsFromDB(collection_name)

	/* Change the collection name to itself */
	collection_name = infra.GetLocalNode().Hostname

	for _, post := range posts {
		err := db.TransferToDb(collection_name, post)
		if err != nil {
			fmt.Println("Error in merging data")
		}
	}
}

func SendColln(collection_name string, destNode *node.Node) {
	posts := db.GetAllPostsFromDB(collection_name)
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(posts)

	fmt.Println("Posts obtained are: ", posts)
	fmt.Println("Data encoded is: ", buf.String())
	/* Kind is assigned KIND_TRANSFER because this data should not be re-replicated */
	infra.SendUnicast(destNode.Hostname, buf.String(), KIND_TRANSFER)
}

func NotifyNodeDies(diedKey string) {
	s := AmISuccessor(diedKey)
	if s {
		MergeCollections(diedKey)
		diedNode, _ := ReplRing.Get(diedKey)
		localNode := infra.GetLocalNode()
		succNode, _, _ := ReplRing.Successor(localNode.Keys[0])
		SendColln(diedNode.Hostname+"-replication", succNode)
	}

	p := AmIPredecessor(diedKey)
	if p {
		pnode, _ := FindNewPrimary(diedKey)
		SendCollnToNewPrimary(pnode)
	}
}

/* if Predecessor => send collection copy to successor */
/* if Successor => add the collection that the Predecessor sent to a new collection */
/* Merge the dead node's collection with its own */

/******************** Helper Functions *******************************/

func StructToString(post db.PetGagPost) (encoded_msg string) {
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

func StringToStruct(data string) (post db.PetGagPost) {
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

func makeVoteMsg(post db.PetGagPost) (voteMsg db.VoteMsg) {

	fmt.Println("Constructing vote message")
	voteMsg = db.VoteMsg{}
	voteMsg.BelongsTo = post.BelongsTo
	voteMsg.DbOp = post.DbOp
	voteMsg.ImageId = post.ObjID
	voteMsg.ImageURL = post.ImageURL
	return voteMsg
}

func makeCommentMsg(post db.PetGagPost) (commentMsg db.AddCommentMsg) {
	commentMsg = db.AddCommentMsg{}
	commentMsg.BelongsTo = post.BelongsTo
	commentMsg.DbOp = post.DbOp
	commentMsg.ImageId = post.ObjID
	commentMsg.ImageURL = post.ImageURL
	commentMsg.UName = post.CommentList[0].UserCName
	commentMsg.Comment = post.CommentList[0].Comment
	return commentMsg
}
