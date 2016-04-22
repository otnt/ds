package main

import (
	"fmt"
	DB "github.com/pshastry/mongoDBintegration"
	//"labix.org/v2/mgo/bson"
)

func main() {
	mongoSession := DB.EstablishSession()
	fmt.Println("Successfully created session")
	var objectID []string
	objectID = make([]string, 1024)

	/* Delete any existing data so as to start afresh */
	DB.DeleteAllFromDB(mongoSession, "petGag")
	DB.DeleteAllFromDB(mongoSession, "petGag1")
	DB.DeleteAllFromDB(mongoSession, "petGag2")

	// Insert fake data
	objectID[0] = (DB.InsertPicture(mongoSession, "https://goo.gl/lc37Hh", "PetGag1", "petGag1", "nil")).Hex()
	//fmt.Println("Inserted the picture \n")
	//fmt.Println("The object id is ", objectID[0])

	//fmt.Println("Fetched from DB")
	//fmt.Println("Image URL is:", image.ImageURL)
	//fmt.Println("UserName is: ", image.UserName)

	objectID[1] = (DB.InsertPicture(mongoSession, "https://goo.gl/V6Ki6x", "PetGag2", "petGag2", "nil")).Hex()
	//fmt.Println("Inserted the picture \n")
	//fmt.Println("The object id is ", objectID[1])

	/*var images []DB.SharedImage
	images = DB.GetOwnDB(mongoSession, "petGag")

	for _, each := range images {
		fmt.Println("Image URL is ", each.ImageURL)
		fmt.Println("UserName is: ", each.UserName)
		fmt.Println("The ID is: ", each.ImageID.Hex())
		fmt.Println("The number of votes is: ", each.UpVote)
	}*/

	/*	image := DB.GetFromDB(mongoSession, bson.ObjectIdHex(objectID[0]), "petGag")
		DB.CommentOnPicture(mongoSession, image.ImageID, "friend1", "So cute", "petGag")
		image.Commt = []DB.Comments{{"", ""}}
		image = DB.GetFromDB(mongoSession, bson.ObjectIdHex(objectID[0]), "petGag")
		fmt.Println("Comment on image: ", image.Commt[0].Comment)
		fmt.Println("Comment on image: ", image.Commt[0].UserName) */

	var images DB.SharedImages
	images = DB.GetAllfromDB(mongoSession)
	for i := 0; i < len(images.Images); i++ {
		fmt.Println("Image URL is ", images.Images[i].ImageURL)
		fmt.Println("UserName is: ", images.Images[i].UserName)
		fmt.Println("The ID is: ", images.Images[i].ImageID.Hex())
		fmt.Println("The number of votes is: ", images.Images[i].UpVote)
	}

}
