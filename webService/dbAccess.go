package webService

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"log"
)

type Comments struct {
	UserCName 		string			`bson: "user_cname"`
	Comment         string			`bson: "user_comment"`
}

type PetGagPost struct {
	UserId			bson.ObjectId	`bson: "_id,omitempty"`
	UserName		string			`bson: "user_name"`
	ImageUrl		string			`bson: "image_url"`
	//	ImageFile		binData			`bson: "image_file"`
	UpVotes			int				`bson: "up_votes"`
	DownVotes		int				`bson: "down_votes"`
	CommentList		[]Comments		`bson: "comments"`
}


func connect() (session *mgo.Session) {
	connectURL := "localhost"
	session, error := mgo.Dial(connectURL)
	if error != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", error)
		os.Exit(1)
	} else {
		//fmt.Printf("Session created\n")
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
	//for _, obj := range allPosts {
	//	fmt.Println(obj)
	//}
	//fmt.Println("Results All: ", allPosts) 
	return allPosts
}

