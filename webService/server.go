package webService

import (
	"encoding/json"
	"github.com/gorilla/mux"
	ch "github.com/otnt/ds/consistentHashing"
	infra "github.com/otnt/ds/infra"
	"net/http"
	"fmt"
	"time"
)

var ring *ch.Ring

const (
	KIND_FORWARD = "forward"
	KIND_FETCH = "fetch"
)

type WebService struct {
	Port int
	router *mux.Router
}

func (ws *WebService) Run(r *ch.Ring) {
	ring = r
	ws.initRouter()
	ws.initHttp()
	ws.initListener()

	block := make(chan bool)
	<-block
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
			case msg := <-infra.ReceivedBuffer:
				if kind := msg.Kind; kind == KIND_FORWARD {
					fmt.Printf("Handle forward message %+v\n", msg)
					infra.SendUnicast(msg.Src, "ok", KIND_FORWARD)
				} else if kind == KIND_FETCH {
					fmt.Println("Fetch all post")
				} else {
					fmt.Println("No support message kind " + kind)
				}
			case <-time.After(time.Millisecond * 1):
				continue
			}
		}
	}()
}

// Fetch all posts in reverse chronological order
func fetchAllPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;  charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(getAllPostsFromDB()); err != nil {
		panic(err)
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
	primary, err := ring.LookUp(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("%q\n", err)))
		return
	}

	//forward message
	fmt.Println("Receive request, forward to " + primary.Hostname)
	infra.SendUnicast(primary.Hostname, np.String(), KIND_FORWARD)
	msg := <-infra.ReceivedBuffer
	fmt.Printf("Receive response %+v\n", msg)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}
