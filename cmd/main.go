package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

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
	name := req.FormValue("name")
	pass := req.FormValue("password")
	redirectTarget := "/login"
	// Just a test
	if name == "ejames" && pass == "temporary" {
		setSession(name, rw)
		redirectTarget = "/admin"
	}
	http.Redirect(rw, req, redirectTarget, 302)
}

func logoutHandler(rw http.ResponseWriter, req *http.Request) {
	clearSession(rw)
	http.Redirect(rw, req, "/admin", 302)
}

func adminHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprint(rw, "Hello private world.")
}

func authMiddleware(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	next(rw, req)
}

/* Helpers */
func setSession(userName string, rw http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}

	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(rw, cookie)
	}
}

func clearSession(rw http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(rw, cookie)
}

func standardMiddleware() *negroni.Negroni {
	return negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
}
