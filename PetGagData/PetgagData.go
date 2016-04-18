package PetGagData

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

type PetgagData struct {
	Data
	BelongsTo string
	DbOp DbOperation
	ImageURL string
	Comment string
	UpVote int
	DownVote int
	UserName string
	ObjID string
}

func NewPetGagData(data Data, dbOp DbOp, comment string, vote int, userName string, objID string, collection string, imageURL string, belongsTo string) (petGagData *PetgagData) {
	petGagData = &PetgagData{}
	petGagData.Data = data
	petGagData.DbOp = dbOp
	petGagData.Comment = comment
	petGagData.Vote = vote
	petGagData.UserName = userName
	petGagData.ObjID = objID
	petGagData.Collection = collection
	petGagData.ImageURL = imageURL
	petGagData.BelongsTo = belongsTo
	return petGagData
}