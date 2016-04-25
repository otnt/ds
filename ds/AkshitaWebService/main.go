package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	router := mux.NewRouter()

	AddHandlers(router)
	http.Handle("/", &EMServer{router})

	fmt.Println("Listening on port 8080")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println(err)
	}
}