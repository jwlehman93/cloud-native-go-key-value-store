package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func helloGoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello net/http!")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", helloGoHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}
