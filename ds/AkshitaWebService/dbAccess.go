package main

import (
	"fmt"
	"github.com/gabstv/go-mgoplus"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"os"
)

/* struct used to insert a comment on a particular post*/
type AddCommentMsg struct {
	BelongsTo string
	DbOp      string
	ImageId   string
	ImageURL  string
	UName     string
	Comment   string
}

/* struct used to upvote/downvote  a particular post*/
type VoteMsg struct {
	BelongsTo string
	ImageId   string
	ImageURL  string
}

type Comments struct {
	UserCName string
	Comment   string
}

type PetGagPost struct {
	/*	PetGagMessage *message.Message */
	BelongsTo   string
	DbOp        string
	ImageURL    string
	CommentList []Comments
	UpVote      int
	DownVote    int
	UserName    string
	ObjID       string
	ImgID       bson.ObjectId `bson:"_id"`
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

func getAllPostsFromDB(collection_name string) (allPosts []bson.M) {
	session := connect()
	defer session.Close()

	collection := session.DB("PetGagDatabase").C(collection_name)

	err := collection.Find(nil).All(&allPosts)
	if err != nil {
		log.Printf("Get from DB error : %s\n", err)
		return
	}
	for _, obj := range allPosts {
		fmt.Println(obj)
	}

	return allPosts
}

func getAllPostsFromAllCollections() (allPosts []bson.M) {
	session := connect()
	defer session.Close()

	db := mongoSession.DB("PetGagDatabase")
	collection_names, err := mgoplus.GetCollectionNames(db)

	if err != nil {
		log.Fatal(err)
	}

	var result bson.M

	for _, each := range collection_names {
		collection := mongoSession.DB("PetGagDatabase").C(each)
		iter := collection.Find(nil).Iter()
		for iter.Next(&result) {
			allPosts = append(allPosts, result)
		}
	}
	return allPosts
}

func (msg *AddCommentMsg) addCommmentInDB() (err error) {
	session := connect()
	defer session.Close()

	collection_name := msg.BelongsTo

	collection := session.DB("PetGagDatabase").C(collection_name)

	id := bson.ObjectIdHex(msg.ImageId)

	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$push": bson.M{"commentlist": bson.M{"usercname": msg.UName, "comment": msg.Comment}}}, ReturnNew: true}
	_, err = doc.Apply(change, &doc)

	if err != nil {
		log.Fatal(err)
	}

	return err

}

func (msg *VoteMsg) downvotePost() (err error) {
	session := connect()
	defer session.Close()

	collection_name := msg.BelongsTo

	collection := session.DB("PetGagDatabase").C(collection_name)

	id := bson.ObjectIdHex(msg.ImageId)

	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"downvote": 1} /*, "$push": bson.M{"SharedImage.$.UpVotedUsers.$.UserName": user_name}*/}, ReturnNew: true}
	_, err = doc.Apply(change, &doc)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (msg *VoteMsg) upvotePost() (err error) {
	session := connect()
	defer session.Close()

	collection_name := msg.BelongsTo
	collection := session.DB("PetGagDatabase").C(collection_name)

	id := bson.ObjectIdHex(msg.ImageId)
	fmt.Println(id)

	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"upvote": 1} /*, "$push": bson.M{"SharedImage.$.UpVotedUsers.$.UserName": user_name}*/}, ReturnNew: true}
	_, err = doc.Apply(change, &doc)

	if err != nil {
		log.Fatal(err)
	}

	return err
}

func (post *PetGagPost) Write() (uid string, err error) { /* Returns objectID in string format along with error */
	//fmt.Println(post)
	session := connect()
	defer session.Close()

	collection_name := post.BelongsTo
	collection := session.DB("PetGagDatabase").C(collection_name)

	if post.ObjID == "nil" {
		i = bson.NewObjectId()
	} else {
		if bson.IsObjectIdHex(post.ObjID) {
			i = bson.ObjectIdHex(post.ObjID)
		} else {
			fmt.Println("Not a valid Object ID")
			i = bson.NewObjectId()
		}
	}

	err := collection.Insert(&PetGagPost{ImageID: i, ImageURL: post.ImageURL, UserName: post.UserName, UpVote: 0, DownVote: 0})
	if err != nil {
		log.Fatal(err)
	}
	return i, err
}
