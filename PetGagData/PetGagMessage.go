package PetGagMessage

import {
	"fmt"
	"Message"
}


/* Enums for the operations that are requested from the client */
type DbOperation int

const {
	Insert DbOperation = iota
	Upvote
	Downvote
	Comment
	Delete
}

type PetGagMessage struct {
	Message
	DbOp string/* Can be insert into DB, upvote, downvote, comment or delete */

}

func NewPetGagMessage(message Message, dbOp string) {
	


}