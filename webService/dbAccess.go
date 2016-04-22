package webService

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"log"
)

/* struct used to insert a comment on a particular post*/
type AddCommentMsg struct {
	ImageURL string
	UName string
	Comment string
}

/* struct used to upvote/downvote  a particular post*/
type VoteMsg struct {
	ImageURL string
}

type Comments struct {
		UserCName 		string			`bson: "UserCName"`
		Comment         string			`bson: "Comment"`
}

type PetGagPost struct {
	/*	PetGagMessage *message.Message */
	BelongsTo     string
	DbOp          string
	ImageURL      string `bson: "ImageUrl"`
	CommentList   []Comments `bson: "CommentList"`
	UpVote        int `bson: "UpVote"`
	DownVote      int `bson: "DownVote"`
	UserName      string `bson: "UserName"`
	UserID        string
	//ObjID         string /* Use same name to avoid confusing */
	ImageId 	  string
}


func connect() (session *mgo.Session) {
	connectURL := "localhost"
	session, error := mgo.Dial(connectURL)
	if error != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", error)
		os.Exit(1)
	} else {
		fmt.Printf("Session created\n")
	}
	session.SetSafe(&mgo.Safe{})
	return session
}

func getAllPostsFromDB() (allPosts []bson.M) {
	session := connect()
	defer session.Close()

	collection := session.DB("PetGagDatabase").C("PetGagPosts")

	err := collection.Find(nil).All(&allPosts)
	if(err != nil){
		log.Printf("Get from DB error : %s\n",err)
		return
	}
	for _, obj := range allPosts {
   		fmt.Println(obj)
  	}

	return allPosts
}

func (msg *AddCommentMsg) addCommmentInDB() (err error) {
	session := connect()
	defer session.Close()

	collection := session.DB("PetGagDatabase").C("PetGagPosts")

	id := bson.ObjectIdHex(msg.ImageURL) 

	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$push": bson.M{"CommentList": bson.M{"UserCName": msg.UName, "Comment": msg.Comment}}}, ReturnNew: true}
	_, err = doc.Apply(change, &doc)


	if err != nil {
		log.Fatal(err)
	}

	return err

}

func (msg *VoteMsg) downvotePost() (err error){
	session := connect()
	defer session.Close()

	collection := session.DB("PetGagDatabase").C("PetGagPosts")

	id := bson.ObjectIdHex(msg.ImageURL) 

	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"DownVote": 1} /*, "$push": bson.M{"SharedImage.$.UpVotedUsers.$.UserName": user_name}*/}, ReturnNew: true}
	_, err = doc.Apply(change, &doc)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (msg *VoteMsg) upvotePost() (err error){
	session := connect()
	defer session.Close()

	collection := session.DB("PetGagDatabase").C("PetGagPosts")

	id := bson.ObjectIdHex(msg.ImageURL)
	fmt.Println(id) 


	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"UpVote": 1} /*, "$push": bson.M{"SharedImage.$.UpVotedUsers.$.UserName": user_name}*/}, ReturnNew: true}
	_, err = doc.Apply(change, &doc)

	if err != nil {
		log.Fatal(err)
	}

	return err
}


func (post *PetGagPost) Write() (err error) {
	//fmt.Println(post)
	session := connect()
	defer session.Close()

	collection := session.DB("PetGagDatabase").C("PetGagPosts")

	// will change following api, inserts the json tags as 'lower case' which is not right!
	err = collection.Insert(post)

	if err != nil {
		log.Fatal(err)
	}
	return err
}
