package main

import (
	"fmt"
	"log"
	"net/http"

	"./api"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/add", api.AddContact).Methods("POST")
	router.HandleFunc("/search", api.Search).Methods("POST")
	router.HandleFunc("/sort", api.Sort).Methods("POST")
	router.HandleFunc("/getall", api.GetAllContacts).Methods("GET")
	fmt.Println("Starting server on port 8000...")
	log.Fatal(http.ListenAndServe(":8000", router))
}
