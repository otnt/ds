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
	BelongsTo     string 
	DbOp          string
	ImageId string
	ImageURL string
	UName string
	Comment string
}

/* struct used to upvote/downvote  a particular post*/
type VoteMsg struct {
	BelongsTo     string 
	DbOp          string
	ImageId string
	ImageURL string
}

type Comments struct {
	UserCName 		string			
	Comment         string
}

type PetGagPost struct {
	/*	PetGagMessage *message.Message */
	BelongsTo     string 
	DbOp          string
	ImageURL      string
	CommentList   []Comments
	UpVote        int
	DownVote      int
	UserName      string
	ObjID         string
}

func (vm *VoteMsg) toPetGagPost() PetGagPost {
	return PetGagPost {
		BelongsTo:vm.BelongsTo,
		DbOp:vm.DbOp,
		ImageURL:vm.ImageURL,
		CommentList:nil,
		UpVote:0,
		DownVote:0,
		UserName:"",
		ObjID:"",
	}
}

func (acm *AddCommentMsg) toPetGagPost() PetGagPost {
	return PetGagPost {
		BelongsTo:acm.BelongsTo,
		DbOp:acm.DbOp,
		ImageURL:acm.ImageURL,
		CommentList:[]Comments{Comments{acm.UName, acm.Comment}},
		UpVote:0,
		DownVote:0,
		UserName:"",
		ObjID:"",
	}
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

	id := bson.ObjectIdHex(msg.ImageId) 

	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$push": bson.M{"commentlist": bson.M{"usercname": msg.UName, "comment": msg.Comment}}}, ReturnNew: true}
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

	id := bson.ObjectIdHex(msg.ImageId) 

	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"downvote": 1} /*, "$push": bson.M{"SharedImage.$.UpVotedUsers.$.UserName": user_name}*/}, ReturnNew: true}
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


func (post *PetGagPost) Write() (err error) {
	//fmt.Println(post)
	session := connect()
	defer session.Close()

	collection := session.DB("PetGagDatabase").C("PetGagPosts")

	// Prathi should do something like : if(post.ObjId is nil and then create an id before inserting into db)
	// simply do post.ObjId = new Object id()
	err = collection.Insert(post)

	if err != nil {
		log.Fatal(err)
	}
	return err
}
