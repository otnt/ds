package main

import (
	"fmt"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

func fetchAllPosts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;  charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(getAllPostsFromDB()); err != nil {
		panic(err)
	}
}

func addPost(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	newPost := &PetGagPost{}
	err = json.Unmarshal(body, &newPost)

    fmt.Println(newPost)

    err = newPost.Write()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("%q\n",err)))
		return
	}

	

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

func addComment(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	msg := &AddCommentMsg{}
	err = json.Unmarshal(body, &msg)

	fmt.Println(msg)

	err = msg.addCommmentInDB()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("%q\n",err)))
		return
	}

	

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))

}

func upVote(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	msg := &VoteMsg{}
	err = json.Unmarshal(body, &msg)

	fmt.Println(msg)

	err = msg.upvotePost()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("%q\n",err)))
		return
	}

	

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))

}

func downVote(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		panic(err)
	}

	msg := &VoteMsg{}
	err = json.Unmarshal(body, &msg)

	fmt.Println(msg)

	err = msg.downvotePost()

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("%q\n",err)))
		return
	}

	

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))

}

type EMServer struct {
	router *mux.Router
}

func (server *EMServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", `POST, GET, OPTIONS, PUT, DELETE`)
		w.Header().Set("Access-Control-Allow-Headers", `Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization`)
	}
	if r.Method == "OPTIONS" {
		return
	}
	server.router.ServeHTTP(w,r)
}

func AddHandlers(router *mux.Router) {
	router.HandleFunc("/fetchAllPosts", fetchAllPosts).Methods("GET")
	router.HandleFunc("/addPost", addPost).Methods("POST")
	router.HandleFunc("/addComment", addComment).Methods("POST")
	router.HandleFunc("/upVote", upVote).Methods("POST")
	router.HandleFunc("/downVote", downVote).Methods("POST")

}