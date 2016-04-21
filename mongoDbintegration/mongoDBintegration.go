package mongoDBintegration

import (
	//"github.com/pshastry/node"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	//"time"
	//"sync"
)

type (
	// Contains information about the comments - this is an embedded document
	Comments struct {
		Comment string `bson:"comments"`
		//UserID   string `bson:"user_comment"`
		UserName string `bson:"user_name"`
	}

	// Contains information about the main document - the image uploaded
	SharedImage struct {
		ImageID  bson.ObjectId `bson:"_id"`
		ImageURL string        `bson:"image_data"`
		UserName string        `bson:"user_name"`
		UpVote   int           `bson:"upvote_num"`
		DownVote int           `bson:"downvote_num"`
		Commt    []Comments    `bson:"comment"`
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
	fmt.Println("Successfully created session")
	return mongoSession

}

/* Get a single image given the objectID */

func GetFromDB(mongoSession *mgo.Session, id bson.ObjectId, collection_name string) SharedImage {
	sessionCopy := mongoSession.Copy()
	defer sessionCopy.Close()

	// Get a collection to execute the query against
	collection := mongoSession.DB("Database").C(collection_name)
	// Retrieve the image
	var image SharedImage
	err := collection.FindId(id).One(&image)
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

/* Get all images posted by all users. Loops through all collections */
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
func InsertPicture(mongoSession *mgo.Session, imageURL string, user_name string, collection_name string, objID string) (i bson.ObjectId) {
	if objID == "nil" {
		i = bson.NewObjectId()
	} else {
		if bson.IsObjectIdHex(objID) {
			i = bson.ObjectIdHex(objID)
		} else {
			fmt.Println("Not a valid Object ID")
			i = bson.NewObjectId()
		}
	}
	//collection := mongoSession.DB("Database").C("SharedImages")
	collection := mongoSession.DB("Database").C(collection_name)
	//err := SharedImages.Insert(image)
	err := collection.Insert(&SharedImage{ImageID: i, ImageURL: imageURL, UserName: user_name, UpVote: 0, DownVote: 0})
	if err != nil {
		log.Println("Insert to DB error : %s\n", err)
	}
	return i
}

/* Increments number of upvotes by vote */
func UpVotePicture(mongoSession *mgo.Session, id bson.ObjectId, vote int, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"upvote_num": vote} /*, "$push": bson.M{"SharedImage.$.UpVotedUsers.$.UserName": user_name}*/}, ReturnNew: true}
	_, err := doc.Apply(change, &doc)
	if err != nil {
		log.Println("Update error : %s\n", err)
	}
}

/* Increments number of downvotes by vote */
func DownVotePicture(mongoSession *mgo.Session, id bson.ObjectId, vote int, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$inc": bson.M{"downvote_num": vote} /*, "$push": bson.M{"SharedImage.$.DownVotedUsers.$.UserName": user_name}*/}, ReturnNew: true}
	_, err := doc.Apply(change, &doc)
	if err != nil {
		log.Println("Update error : %s\n", err)
	}
}

func CommentOnPicture(mongoSession *mgo.Session, id bson.ObjectId, un string, comt string, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	doc := collection.FindId(id)
	change := mgo.Change{Update: bson.M{"$push": bson.M{"comment": bson.M{"user_name": un, "comments": comt}}}, ReturnNew: true}
	_, err := doc.Apply(change, &doc)
	if err != nil {
		log.Println("Update error : %s\n", err)
	}

}

func DeleteFromDB(mongoSession *mgo.Session, id bson.ObjectId, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	err := collection.Remove(bson.ObjectId(id))
	if err != nil {
		log.Println("Remove from DB error : %s\n", err)
		return
	}
}

func DeleteAllFromDB(mongoSession *mgo.Session, collection_name string) {
	collection := mongoSession.DB("Database").C(collection_name)
	_, err := collection.RemoveAll(nil)
	if err != nil {
		log.Println("Remove all from DB error : %s\n", err)
		return
	}
}
