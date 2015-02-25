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
	"github.com/russross/blackfriday"
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
			"fdate": dateFmt,
			"md":    markdown,
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
	admin.HandleFunc("/edit/{link}", editPostDisplayHandler).Methods("GET")
	admin.HandleFunc("/edit/{link}", editPostHandler).Methods("POST")
	admin.HandleFunc("/delete/{link}", deletePostHandler).Methods("DELETE")
	admin.HandleFunc("/regen", regenerateSiteHandler).Methods("POST")

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
	presenter := *blog
	presenter.Posts = blog.GetAllPosts()
	rndr.HTML(rw, http.StatusOK, "admin", presenter)
}

func newPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	rndr.HTML(rw, http.StatusOK, "newpost", blog)
}

func newPostHandler(rw http.ResponseWriter, req *http.Request) {
	post := &goblawg.Post{}
	post.Body = []byte(req.FormValue("body"))
	post.Title = req.FormValue("title")
	post.Link = goblawg.LinkifyTitle(post.Title)

	isDraft := false
	if req.FormValue("draft") == "true" {
		isDraft = true
	}
	post.IsDraft = isDraft

	// Change this later
	post.Time = time.Now()

	post.LastModified = time.Now()

	err := blog.SavePost(post)
	// TODO: Change to session to display error.
	if err != nil {
		fmt.Fprintln(rw, "Post save error, %v", err)
		return
	}

	http.Redirect(rw, req, "/admin", 302)
}

func editPostDisplayHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	link := vars["link"]
	post := blog.GetPostByLink(link)

	presenter := struct {
		Name         string
		BlogLink     string
		Title        string
		Body         string
		Link         string
		Time         time.Time
		IsDraft      bool
		LastModified time.Time
	}{
		blog.Name,
		blog.Link,
		post.Title,
		string(post.Body),
		post.Link,
		post.Time,
		post.IsDraft,
		post.LastModified,
	}

	rndr.HTML(rw, http.StatusOK, "edit", presenter)
}

func editPostHandler(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	link := vars["link"]
	post := blog.GetPostByLink(link)

	fmt.Println(post)
}

func deletePostHandler(rw http.ResponseWriter, req *http.Request) {
	link := mux.Vars(req)["link"]
	post := blog.GetPostByLink(link)
	blog.DeletePost(post)

	rndr.JSON(rw, http.StatusNoContent, nil)
}

func regenerateSiteHandler(rw http.ResponseWriter, req *http.Request) {
	err := blog.GenerateSite()

	if err != nil {
		fmt.Fprintf(rw, "ERROR: %v", err)
		return
	}

	fmt.Fprintf(rw, "Success!")
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

func markdown(input []byte) string {
	output := blackfriday.MarkdownCommon(input)
	return string(output)
}
