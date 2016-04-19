package mongoDBintegration

import (
	//"github.com/pshastry/node"
	//"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	//"time"
	//"sync"
)

/* const {
	MongoDBHosts = ""
	AuthDB = ""
	AuthUserName = ""
	AuthPassword = ""
} */

type (
	// Contains information about the comments - this is an embedded document
	Comments struct {
		Comment string `bson:"comment"`
		//UserID   string `bson:"user_comment"`
		UserName string `bson:"user_name"`
	}

	// Contains information about the votes - this is an embedded document
	UpVotes struct {
		//UpVote int 'bson: "upvote_num"'
		//UserID   string `bson:"upvoter_id"`
		//Upvotes  int    `bson:"upvote_num"`
		UpvotedUsers []string `bson:"user_name"`
	}

	DownVotes struct {
		//Downvotes int    `bson:"downvote_num"`
		DownvotedUsers []string `bson:"user_name"`
	}

	// Contains information about the main document - the image uploaded
	SharedImage struct {
		ImageID        bson.ObjectId `bson:"_id, omitempty"`
		ImageURL       string        `bson:"image_data"`
		UserName       string        `bson:"user_name"`
		UpVote         int           `bson:"upvote_num"`
		DownVote       int           `bson:"downvote_num"`
		Commt          []Comments    `bson:"comment"`
		UpVotedUsers   UpVotes       `bson:"upvote"`
		DownVotedUsers DownVotes     `bson:"downvote"`
	}

	SharedImages struct {
		Images []SharedImage
	}
)

// Establish a session with our mongoDB database
func EstablishSession() *mgo.Session {
	/*	mongoDBDialInfo := &mgo.DialInfo { Addrs: []string{MongoDBHosts},
								Timeout: 60*time.Second, Database: AuthDB,
								Username: AuthUserName,
								Password: AuthPassword}
		mongoSession, error := mgo.DialWithInfo(mongoDBDialInfo)
	*/
	// Create a new session
	host := "localhost"
	mongoSession, err := mgo.Dial(host)
	if err != nil {
		log.Fatalf("Create session: %s\n", err)
	}

	/* SetMode changes the consistency mode for the session.
	In the Monotonic consistency mode reads may not be entirely up-to-date,
	but they will always see the history of changes moving forward, the data
	read will be consistent across sequential queries in the same session, and
	modifications made within the session will be observed in following queries
	(read-your-write)
	If refresh is true, in addition to ensuring the session is in the given
	consistency mode, the consistency guarantees will also be reset
	(e.g. a Monotonic session will be allowed to read from secondaries again).*/

	mongoSession.SetMode(mgo.Monotonic, true)
	return mongoSession

}

/* Get a single image given the objectID */

func GetFromDB(mongoSession *mgo.Session, id bson.ObjectId) SharedImage {
	sessionCopy := mongoSession.Copy()
	defer sessionCopy.Close()

	// Get a collection to execute the query against
	collection := mongoSession.DB("Database").C("SharedImages")
	// Retrieve the image
	var image SharedImage
	err := collection.FindId(bson.ObjectId(id)).One(&image)
	if err != nil {
		log.Println("Get from DB error : %s\n", err)
	}
	return image

}

/* Get all images posted by a single user */
func GetOwnDB(mongoSession *mgo.Session, collection_name string) []SharedImage {
	sessionCopy := mongoSession.Copy()
	defer sessionCopy.Close()

	collection := mongoSession.DB("Database").C(collection_name)
	var images []SharedImage
	err := collection.Find(nil).All(&images)
	if err != nil {
		log.Println("Get from DB error : %s\n", err)
	}
	return images
}

/* Get all images posted by all users */
func GetAllfromDB(mongoSession *mgo.Session) SharedImages {
	sessionCopy := mongoSession.Copy()
	defer sessionCopy.Close()

	var collection_names []string
	var err error
	collection_names, err = mongoSession.DB("Database").CollectionNames()
	if err != nil {
		log.Println("Error in getting collection names : %s\n", err)
	}
	var imagesAll SharedImages
	var result SharedImage
	for _, each := range collection_names {
		collection := mongoSession.DB("Database").C(each)
		iter := collection.Find(nil).Iter()
		for iter.Next(&result) {
			imagesAll.Images = append(imagesAll.Images, result)
		}
	}
	return imagesAll
}

/* Max file size supported is 16 MB */
func InsertPicture(mongoSession *mgo.Session, imageURL string, user_name string, collection_name string) (i bson.ObjectId) {
	i = bson.NewObjectId()
	//collection := mongoSession.DB("Database").C("SharedImages")
	collection := mongoSession.DB("Database").C(collection_name)
	//err := SharedImages.Insert(image)
	err := collection.Insert(&SharedImage{ImageID: i, ImageURL: imageURL, UserName: user_name, UpVote: 0, DownVote: 0})
	if err != nil {
		log.Println("Insert to DB error : %s\n", err)
	}
	return i
}

func UpVotePicture(mongoSession *mgo.Session, id bson.ObjectId, user_name string, vote int, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"SharedImage.$.UpVote": 1}, "$push": bson.M{"SharedImage.$.UpVotedUsers.$.UserName": user_name}}, ReturnNew: true}
	_, err := doc.Apply(change, &doc)
	if err != nil {
		log.Println("Update error : %s\n", err)
	}
}

func DownVotePicture(mongoSession *mgo.Session, id bson.ObjectId, user_name string, vote int, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"SharedImage.$.DownVote": 1}, "$push": bson.M{"SharedImage.$.DownVotedUsers.$.UserName": user_name}}, ReturnNew: true}
	_, err := doc.Apply(change, &doc)
	if err != nil {
		log.Println("Update error : %s\n", err)
	}
}

func CommentOnPicture(mongoSession *mgo.Session, id bson.ObjectId, user_name string, comt string, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$push": bson.M{"SharedImage.$.Commt.$.UserName": user_name, "SharedImage.$.Commt.$.Comment": comt}}, ReturnNew: true}
	_, err := doc.Apply(change, &doc)
	if err != nil {
		log.Println("Update error : %s\n", err)
	}

}

func DeleteFromDB(mongoSession *mgo.Session, id bson.ObjectId, collection_name string) {
	collection := mongoSession.DB("Database").C("collection_name")
	err := collection.Remove(bson.ObjectId(id))
	if err != nil {
		log.Println("Remove from DB error : %s\n", err)
		return
	}

}
