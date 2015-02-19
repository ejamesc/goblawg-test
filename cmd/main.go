package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/codegangsta/negroni"
	"github.com/ejamesc/goblawg"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/unrolled/render"
)

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

var rndr = render.New(render.Options{
	Directory:  "templates",
	Extensions: []string{".html"},
	Layout:     "base",
	Funcs: []template.FuncMap{
		template.FuncMap{
			"fdate":     dateFmt,
			"sortPosts": goblawg.SortPosts,
		},
	},
})

var blog *goblawg.Blog

/*
 * Main Function
 */
func main() {
	settings, err := ioutil.ReadFile("settings.json")

	if err != nil {
		fmt.Println("Error with reading settings: %s", err)
	}

	blog, err = goblawg.NewBlog(string(settings))

	if err != nil {
		fmt.Printf("Error with creating new blog: %s\n", err)
	}

	/* Set up middleware */

	r := mux.NewRouter()

	adminBase := mux.NewRouter()
	adminBase.HandleFunc("/admin", adminHandler)
	r.PathPrefix("/admin").Handler(
		negroni.New(negroni.HandlerFunc(authMiddleware),
			negroni.Wrap(adminBase),
		))
	admin := adminBase.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/new", newPostDisplayHandler).Methods("GET")
	admin.HandleFunc("/new", newPostHandler).Methods("POST")

	/* Global Routes */
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("static"))))

	// r.Handle("/", http.FileServer(http.Dir(blog.OutDir)))
	r.HandleFunc("/login", loginDisplayHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/logout", logoutHandler).Methods("POST")

	n := standardMiddleware()
	n.UseHandler(r)
	n.Run(":3000")
}

func loginDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	if getUserName(req) == "ejames" {
		http.Redirect(rw, req, "/admin", 302)
	} else {
		rndr.HTML(rw, http.StatusOK, "login", blog)
	}
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
	http.Redirect(rw, req, "/login", 302)
}

func adminHandler(rw http.ResponseWriter, req *http.Request) {
	rndr.HTML(rw, http.StatusOK, "admin", blog)
}

func newPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	rndr.HTML(rw, http.StatusOK, "newpost", blog)
}

func newPostHandler(rw http.ResponseWriter, req *http.Request) {
	post := &goblawg.Post{}
	post.Body = []byte(req.FormValue("body"))
	post.Title = req.FormValue("title")
	link := ""
	if req.FormValue("link") != "" {
		link = req.FormValue("link")
	} else {
		link = goblawg.LinkifyTitle(post.Title)
	}
	post.Link = link

	isDraft := false
	if req.FormValue("draft") == "true" {
		isDraft = true
	}
	post.IsDraft = isDraft

	// Change this later
	post.Time = time.Now()

	post.LastModified = time.Now()

	err := blog.SavePost(post)
	// Change to session to display error.
	if err != nil {
		fmt.Fprintln(rw, "Post save error, %v", err)
	}

	fmt.Fprintln(rw, "New post successfully created")
}

	fmt.Fprintln(rw, "NEW POST")
}

func authMiddleware(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	if getUserName(req) == "ejames" {
		next(rw, req)
	} else {
		http.Redirect(rw, req, "/login", 302)
	}
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

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func standardMiddleware() *negroni.Negroni {
	return negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger())
}

/* Template functions */

func dateFmt(tt time.Time) string {
	const layout = "3:04pm, 2 January 2006"
	return tt.Format(layout)
}
