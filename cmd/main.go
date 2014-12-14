package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
)

/*
 * Main Function
 */
func main() {
	settings, err := ioutil.ReadFile("settings.json")

	if err != nil {
		fmt.Println("Error with reading settings: %s", err)
	}

	_, err = goblawg.NewBlog(string(settings))

	if err != nil {
		fmt.Printf("Error with creating new blog: %s\n", err)
	}

	/* Set up middleware */

	r := mux.NewRouter()

	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/", adminHandler)

	an := negroni.New(negroni.HandlerFunc(authMiddleware))
	an.UseHandler(adminRouter)

	/* Global Routes */
	r.Handle("/admin", an)
	// r.Handle("/", http.FileServer(http.Dir(blog.OutDir)))
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("POST")

	n := standardMiddleware()
	n.UseHandler(r)
	n.Run(":3000")
}

func loginHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "Hello login handler.")
}

func logoutHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "Hello logout handler")
}

func adminHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "Hello private world.")
}

func authMiddleware(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	next(rw, req)
}

/* Helpers */
func standardMiddleware() *negroni.Negroni {
	return negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
}
