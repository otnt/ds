package webService

import (
	"encoding/json"
	"github.com/gorilla/mux"
	ch "github.com/otnt/ds/consistentHashing"
	infra "github.com/otnt/ds/infra"
	"net/http"
	"fmt"
	"time"
	"bytes"
	"github.com/otnt/ds/message"
	"github.com/otnt/ds/node"
	"labix.org/v2/mgo/bson"
	"io/ioutil"
	"errors"
)

var ring *ch.Ring

const (
	KIND_FORWARD = "ws_forward"
	KIND_FETCH = "ws_fetch"
	KIND_FORWARD_ACK = "ws_forward_ack"
	KIND_FETCH_ACK = "ws_fetch_ack"
	KIND_COMMENT = "ws_comment"
	KIND_COMMENT_ACK = "ws_comment_ack"
	KIND_UP_VOTE = "ws_up_vote"
	KIND_UP_VOTE_ACK = "ws_up_vote_ack"
	KIND_DOWN_VOTE = "ws_down_vote"
	KIND_DOWN_VOTE_ACK = "ws_down_vote_ack"
)

const (
	TIME_OUT = time.Millisecond * 3000
)

var	ForwardChan chan *message.Message
var	FetchChan chan *message.Message
var	ForwardAckChan chan *message.Message
var	FetchAckChan chan *message.Message
var CommentChan chan *message.Message
var CommentAckChan chan *message.Message
var UpVoteChan chan *message.Message
var UpVoteAckChan chan *message.Message
var DownVoteChan chan *message.Message
var DownVoteAckChan chan *message.Message

type WebService struct {
	Port int
	router *mux.Router
}

func (ws *WebService) Run(r *ch.Ring) {
	ForwardChan = make(chan *message.Message)
	FetchChan= make(chan *message.Message)
	ForwardAckChan= make(chan *message.Message)
	FetchAckChan= make(chan *message.Message)
	CommentChan = make(chan *message.Message)
	CommentAckChan = make(chan *message.Message)
	UpVoteChan = make(chan *message.Message)
	UpVoteAckChan = make(chan *message.Message)
	DownVoteChan = make(chan *message.Message)
	DownVoteAckChan = make(chan *message.Message)
	ring = r
	ws.initRouter()
	ws.initHttp()
	ws.initListener()
}

// Create router and serve the request
func (ws *WebService) initRouter() {
	ws.router = mux.NewRouter()
	ws.router.HandleFunc("/fetchAllPosts", fetchAllPosts).Methods("GET")
	ws.router.HandleFunc("/addPost", createNewPost).Methods("POST")
	ws.router.HandleFunc("/addComment", addComment).Methods("POST")
	ws.router.HandleFunc("/upVote", upVote).Methods("POST")
	ws.router.HandleFunc("/downVote", downVote).Methods("POST")
}

// Create HTTP listener
func (ws *WebService) initHttp() {
	go func() {
		http.Handle("/", ws.router)
		err := http.ListenAndServe(fmt.Sprintf(":%d", ws.Port), nil)
		if err != nil {
			panic(err)
		}
	}()
}

// Listen to incoming forward request, to save data into local db
func (ws *WebService) initListener() {
	go func() {
		for {
			select {
			case msg := <-ForwardChan:
				ws.HandleForward(msg)
			case msg := <-FetchChan:
				ws.HandleFetch(msg)
			case msg := <-CommentChan:
				ws.HandleComment(msg)
			case msg := <-UpVoteChan:
				ws.HandleUpVote(msg)
			case msg := <-DownVoteChan:
				ws.HandleDownVote(msg)
			}
		}
	}()
}

func (ws *WebService) HandleComment(msg *message.Message) {
	fmt.Println("Handle comment from " + msg.Src)

	comment := &AddCommentMsg{}
	err := json.Unmarshal([]byte(msg.Data), &comment)
	if err != nil {
		return
	}

	err = comment.addCommmentInDB()
	if err != nil {
		return
	}

	infra.SendUnicast(msg.Src, "ok", KIND_COMMENT_ACK)
	fmt.Println(comment)
}

func (ws *WebService) HandleForward(msg *message.Message) {
	fmt.Println("Handle forward message from " + msg.Src)
	newPost := &PetGagPost{}
	err := json.Unmarshal([]byte(msg.Data), &newPost)
	if err != nil {
		return
	}

	err = newPost.Write()
	if err != nil {
		return
	}

	infra.SendUnicast(msg.Src, "ok", KIND_FORWARD_ACK)
	fmt.Println(newPost)
}

func (ws *WebService) HandleUpVote(msg *message.Message) {
	fmt.Println("Handle upvote from " + msg.Src)
	vote := &VoteMsg{}
	err := json.Unmarshal([]byte(msg.Data), &vote)
	if err != nil {
		return
	}

	err = vote.upvotePost()
	if err != nil {
		return
	}

	infra.SendUnicast(msg.Src, "ok", KIND_UP_VOTE_ACK)
	fmt.Println(vote)
}

func (ws *WebService) HandleDownVote(msg *message.Message) {
	fmt.Println("Handle downvote from " + msg.Src)
	vote := &VoteMsg{}
	err := json.Unmarshal([]byte(msg.Data), &vote)
	if err != nil {
		return
	}

	err = vote.downvotePost()
	if err != nil {
		return
	}

	infra.SendUnicast(msg.Src, "ok", KIND_DOWN_VOTE_ACK)
	fmt.Println(vote)
}

// Handle fetch request, return all data in local database
func (ws *WebService) HandleFetch(msg *message.Message) {
	fmt.Println("Do fetching")

	posts := getAllPostsFromDB()
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(posts)
	infra.SendUnicast(msg.Src, buf.String(),KIND_FETCH_ACK)
}

// Fetch all posts in reverse chronological order
func fetchAllPosts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receive fetch request")

	for hostname, _ := range infra.NodeIndexMap {
		infra.SendUnicast(hostname, "fetch", KIND_FETCH)
	}

	res := make([]bson.M, 0)
	for _, _ = range infra.NodeIndexMap {
		msg := <-FetchAckChan
		var d []bson.M
		err := json.NewDecoder(bytes.NewBufferString(msg.Data)).Decode(&d)
		if err != nil {
			fmt.Printf("Error when decode data %v\n", err)
			return
		}
		res = append(res, d...)
	}

	fmt.Println("Fetched all data")

	w.Header().Set("Content-Type", "application/json;  charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		panic(fmt.Sprintf("Error when encoding posts: %+v\n", err))
	}
}

// New post request
type newPost struct {
	ImageData string
}

func (np *newPost) String() string {
	return fmt.Sprintf("%s", np.ImageData)
}


// Create a new post.
// It forwards the request to primary node of the coming post.
func createNewPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Add new post")
	body, err := ioutil.ReadAll(r.Body)

	if err == nil {
		newPost := &PetGagPost{}
		err = json.Unmarshal(body, &newPost)

		if err == nil {
			id := newPost.ImageURL

			if id == "" {
				badRequest(w, errors.New("Missing ImageURL"))
				return
			}
			primary, _:= primaryNode(id)
			infra.SendUnicast(primary.Hostname, string(body), KIND_FORWARD)
			ok := waitFor(ForwardAckChan, TIME_OUT)

			if ok {
				requestOk(w)
				return
			} else {
				internalError(w, err)
				return
			}
		}
	}
	badRequest(w, err)
}

// Add a new comment
func addComment(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Add new comment")
	body, err := ioutil.ReadAll(r.Body)

	if err == nil {
		msg := &AddCommentMsg{}
		err = json.Unmarshal(body, &msg)

		if err == nil {
			id := msg.ImageURL

			if id == "" {
				badRequest(w, errors.New("Missing ImageURL"))
				return
			}
			primary, _:= primaryNode(id)
			fmt.Println(primary.Hostname, KIND_COMMENT)
			infra.SendUnicast(primary.Hostname, string(body),KIND_COMMENT)
			ok := waitFor(CommentAckChan, TIME_OUT)

			if ok {
				requestOk(w)
				return
			} else {
				internalError(w, err)
				return
			}
		}
	}
	badRequest(w, err)
}

// Add new up vote
func upVote(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Add new upVote")
	body, err := ioutil.ReadAll(r.Body)

	if err == nil {
		vote := &VoteMsg{}
		err = json.Unmarshal(body, &vote)

		if err == nil {
			id := vote.ImageURL

			if id == "" {
				badRequest(w, errors.New("Missing ImageURL"))
				return
			}
			primary, _:= primaryNode(id)
			infra.SendUnicast(primary.Hostname, string(body),KIND_UP_VOTE)
			ok := waitFor(UpVoteAckChan, TIME_OUT)

			if ok {
				requestOk(w)
				return
			} else {
				internalError(w, err)
				return
			}
		}
	}
	badRequest(w, err)
}

// Add new down vote
func downVote(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Add new downVote")
	body, err := ioutil.ReadAll(r.Body)

	if err == nil {
		vote := &VoteMsg{}
		err = json.Unmarshal(body, &vote)

		if err == nil {
			id := vote.ImageURL

			if id == "" {
				badRequest(w, errors.New("Missing ImageURL"))
				return
			}
			primary, _:= primaryNode(id)
			infra.SendUnicast(primary.Hostname, string(body), KIND_DOWN_VOTE)
			ok := waitFor(DownVoteAckChan, TIME_OUT)

			if ok {
				requestOk(w)
				return
			} else {
				internalError(w, err)
				return
			}
		}
	}
	badRequest(w, err)
}

// look for primary node
func primaryNode(id string) (*node.Node, error) {
	primary, _, err := ring.LookUp(ring.Hash(id))
	if err != nil {
		return nil, err
	}
	return primary, nil
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(fmt.Sprintf("bad request %+v\n", err)))
	return
}

func internalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("server internal error %+v\n", err)))
	return
}

func requestOk(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
	return
}

func waitFor(c <-chan *message.Message, t time.Duration) bool {
	select {
	case <-c:
		return true
	case <-time.After(t):
		return false
	}
}
