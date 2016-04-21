package petgagData

import (
//"github.com/otnt/ds/message"
//"github.com/otnt/ds/petGagMessage"
)

/* type DbOperation int

const (
	OP_INSERT   = "Insert"
	OP_UPVOTE   = "Upvote"
	OP_DOWNVOTE = "Downvote"
	OP_COMMENT  = "Comment"
	OP_DELETE   = "Delete"
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
	ObjID     string /* hex string representation of Object ID. For new item, ObjID = "" */
}

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
