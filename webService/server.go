package webService

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
)

type WebService struct {
	Port int
	router *mux.Router
}

func (ws *WebService) Run() {
	ws.initRouter()
	ws.initHttp()
}

// Create router and serve the request
func (ws *WebService) initRouter() {
	ws.router = mux.NewRouter()
	ws.router.HandleFunc("/fetchAllPosts", fetchAllPosts).Methods("GET")
	ws.router.HandleFunc("/post", createNewPost).Methods("POST")
}

func (ws *WebService) initHttp() {
	// Create HTTP listener
	http.Handle("/", ws.router)
	err := http.ListenAndServe(fmt.Sprintf(":%d", ws.Port), nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Listening on port %d\n", ws.Port)
	}
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

// Create a new post.
// It forwards the request to primary node of the coming post.
func createNewPost(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Not implmented yet"))
	w.WriteHeader(http.StatusNotImplemented)
}
