package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

/*
 * Main Function
 */
func main() {
	r := mux.NewRouter()

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
