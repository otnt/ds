package petgagData

import (
//"github.com/otnt/ds/message"
//"github.com/otnt/ds/petGagMessage"
)

/*
type DbOperation int

const (
	Insert DbOperation = 1 + iota
	Upvote
	Downvote
	Comment
	Delete
) */

type PetgagData struct {
	/*	PetGagMessage *message.Message */
	BelongsTo string
	DbOp      string
	ImageURL  string
	Comment   string
	UpVote    int
	DownVote  int
	UserName  string
	UserID    string
	ObjID     string
}

/*func NewPetGagData(msg message.Message, dbOp DbOperation, comment string, upVote int, downVote int, userName string, objID string, collection string, imageURL string, belongsTo string) (petGagData *PetgagData) {
	petGagData = &PetgagData{}
	petGagData.PetGagMessage.Data = msg.Data
	petGagData.DbOp = dbOp
	petGagData.Comment = comment
	petGagData.UpVote = upVote
	petGagData.DownVote = downVote
	petGagData.UserName = userName
	petGagData.ObjID = objID
	petGagData.ImageURL = imageURL
	petGagData.BelongsTo = belongsTo
	return petGagData
}  */

func NewPetGagData(dbOp string, comment string, upVote int, downVote int, userName string, userID string, objID string, collection string, imageURL string, belongsTo string) (petGagData *PetgagData) {
	petGagData = &PetgagData{}
	petGagData.DbOp = dbOp
	petGagData.Comment = comment
	petGagData.UpVote = upVote
	petGagData.DownVote = downVote
	petGagData.UserName = userName
	petGagData.UserID = userID
	petGagData.ObjID = objID
	petGagData.ImageURL = imageURL
	petGagData.BelongsTo = belongsTo
	return petGagData
}
