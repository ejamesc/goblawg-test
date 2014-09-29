package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

/*
 * Structs
 */
type Post struct {
	Title string
	Body  []byte
}

type Content struct {
	Title template.HTML
	Body  template.HTML
}

func (p *Post) save() error {
	filename := convertToMarkdownFilename(p.Title)
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/*
* Utils
 */
func convertToMarkdownFilename(title string) string {
	const layout = "2 Jan 2006 "
	t := time.Now()
	title = t.Format(layout) + title

	res := strings.Replace(title, "-", "_", -1)
	res = strings.Replace(res, " ", "_", -1)
	res = strings.ToLower(res)

	res = "content/" + res + ".md"
	return res
}

func convertToDisplayTitle(urlPath string) string {
	res := strings.Replace(urlPath, "-", " ", -1)
	res = strings.Replace(res, "_", " ", -1)
	res = strings.Title(res)
	return res
}

func saveHTML(title string, template string) error {
	content, err := loadPost(title)
	if err != nil {
		return err
	}

	filepath := strings.Replace(title, "-", "_", -1)
	filepath = strings.Replace(filepath, " ", "_", -1)
	filepath = strings.ToLower(filepath)
	filepath = filepath + ".html"

	w, werr := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	defer w.Close()
	if werr != nil {
		return werr
	}

	templates.ExecuteTemplate(w, template+".html", content)
	return nil
}

func loadPost(title string) (*Content, error) {
	filename := convertToMarkdownFilename(title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	htmlBody := blackfriday.MarkdownCommon(body)
	displayTitle := convertToDisplayTitle(title)
	return &Content{Title: template.HTML(displayTitle), Body: template.HTML(htmlBody)}, nil
}

/*
 * Template Loading
 */
// TODO: better way to do this?
var templates = template.Must(template.ParseFiles(
	"templates/partials.html",
	"templates/admin.html",
	"templates/view.html"))

/*
 * Main Function
 */
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/admin", adminHandler)
	r.HandleFunc("/admin/save", saveHandler).Methods("POST")
	r.HandleFunc("/view/{title}", viewHandler)
	r.HandleFunc("/generate/{title}", generateHandler)
	r.PathPrefix("/static/").Handler(http.FileServer(http.Dir(".")))

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

/*
 * Handlers
 */
// TODO: List out all posts

func adminHandler(w http.ResponseWriter, request *http.Request) {
	renderTemplate(w, "admin", nil)
}

func saveHandler(w http.ResponseWriter, req *http.Request) {
	title := req.FormValue("Title")
	body := req.FormValue("Body")
	post := &Post{Title: title, Body: []byte(body)}

	err := post.save()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/admin", 303)
}

func viewHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	post, err := loadPost(vars["title"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "view", post)
}

func generateHandler(w http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	err := saveHTML(vars["title"], "view")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "Save complete!")
}

func renderTemplate(w http.ResponseWriter, tmpl string, c *Content) {
	err := templates.ExecuteTemplate(w, tmpl+".html", c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
