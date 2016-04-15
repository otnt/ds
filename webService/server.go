package webService

import (
	"encoding/json"
	"github.com/gorilla/mux"
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
}
