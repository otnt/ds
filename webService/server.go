package webService

import (
	"encoding/json"
	"github.com/gorilla/mux"
	ch "github.com/otnt/ds/consistentHashing"
	infra "github.com/otnt/ds/infra"
	"net/http"
	"fmt"
	//"time"
	"gopkg.in/mgo.v2/bson"
	"bytes"
	"github.com/otnt/ds/message"
)

var ring *ch.Ring

const (
	KIND_FORWARD = "forward"
	KIND_FETCH = "fetch"
	KIND_FORWARD_ACK = "forward_ack"
	KIND_FETCH_ACK = "fetch_ack"
)

var	ForwardChan chan *message.Message
var	FetchChan chan *message.Message
var	ForwardAckChan chan *message.Message
var	FetchAckChan chan *message.Message

type WebService struct {
	Port int
	router *mux.Router
}

func (ws *WebService) Run(r *ch.Ring) {
	ForwardChan = make(chan *message.Message, 10)
	FetchChan= make(chan *message.Message, 10)
	ForwardAckChan= make(chan *message.Message, 10)
	FetchAckChan= make(chan *message.Message, 10)
	ring = r
	ws.initRouter()
	ws.initHttp()
	ws.initListener()
}

// Create router and serve the request
func (ws *WebService) initRouter() {
	ws.router = mux.NewRouter()
	ws.router.HandleFunc("/fetchAllPosts", fetchAllPosts).Methods("GET")
	ws.router.HandleFunc("/post", createNewPost).Methods("POST")
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
				fmt.Println("Handle forward message from " + msg.Src)
				infra.SendUnicast(msg.Src, "ok", KIND_FORWARD_ACK)
			case msg := <-FetchChan:
				ws.HandleFetch(msg)
			}
		}
	}()
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
	// get post data
	var np newPost
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&np)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("%q\n", err)))
		return
	}

	// look for primary node
	var length int
	if len(np.ImageData) > 100 {
		length = 100
	} else {
		length = len(np.ImageData)
	}
	data := np.ImageData[:length]
	primary, _, err := ring.LookUp(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("%q\n", err)))
		return
	}

	//forward message
	fmt.Println("Receive request, forward to " + primary.Hostname)
	infra.SendUnicast(primary.Hostname, np.String(), KIND_FORWARD)
	msg := <-ForwardAckChan
	fmt.Println("Receive response from " + msg.Src)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}
