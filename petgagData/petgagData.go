package petgagData

import (
//"github.com/otnt/ds/message"
//"github.com/otnt/ds/petGagMessage"
)

type Comment struct {
	Comt string `bson:"comment"`
	//UserID   string `bson:"user_comment"`
	UserName string `bson:"user_name"`
}

type PetgagData struct {
	/*	PetGagMessage *message.Message */
	BelongsTo string
	DbOp      string /* can take "Insert", "Upvote", "Downvote"", "Comment" or "Delete" values */
	ImageURL  string
	Commt     []Comment
	UpVote    int
	DownVote  int
	UserName  string
	UserID    string
	ObjID     string /* hex string representation of Object ID. For new item, ObjID = "nil" */
}

func NewComment(comt string, userName string) (comment Comment) {
	comment.Comt = comt
	comment.UserName = userName
	return comment
}

func NewPetGagData(dbOp string, comment Comment, upVote int, downVote int, userName string, userID string, objID string, imageURL string, belongsTo string) (data *PetgagData) {
	data = &PetgagData{}
	data.Commt = make([]Comment, 1)
	data.Commt[0] = comment
	//data.Commt.Comment = comment.Comment
	//data.Commt.UserName = comment.UserName
	data.DbOp = dbOp
	data.UpVote = upVote
	data.DownVote = downVote
	data.UserName = userName
	data.UserID = userID
	data.ObjID = objID
	data.ImageURL = imageURL
	data.BelongsTo = belongsTo
	return data
}
