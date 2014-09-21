package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
  "strings"
  "os"
  "fmt"

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
	filename := p.Title + ".md"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

/*
 * Utils
 */
func convertToMarkdownFilename(urlPath string) string {
  res := strings.Replace(urlPath, "-", "_", -1)
  res = "content/" + res + ".md"
  return res
}

func convertToFilename(urlPath string) string {
  res := strings.Replace(urlPath, "-", "_", -1)
  res = strings.Replace(urlPath, " ", "_", -1)
  res = res + ".html"
  return res
}

func convertToDisplayTitle(urlPath string) string {
  res := strings.Replace(urlPath, "-", " ", -1)
  res = strings.Title(res)
  return res
}

func saveHTML(title string, template string) error {
  content, err := loadPost(title)
  if err != nil {
    return err
  }

  filepath := convertToFilename(title)
  w, werr := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
  if werr != nil {
    return werr
  }
  defer w.Close()
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
	r.HandleFunc("/admin/", adminHandler)
	r.HandleFunc("/view/{title}", viewHandler)
  r.HandleFunc("/save/{title}", saveHandler)

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

/*
 * Handlers
 */
// TODO: List out all posts

func adminHandler(writer http.ResponseWriter, request *http.Request) {
	renderTemplate(writer, "admin", nil)
}

func viewHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	post, err := loadPost(vars["title"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(writer, "view", post)
}

func saveHandler(writer http.ResponseWriter, request *http.Request) {
  vars := mux.Vars(request)
  err := saveHTML(vars["title"], "view")
  if err != nil {
    http.Error(writer, err.Error(), http.StatusInternalServerError)
    return
  }

  fmt.Fprint(writer, "Save complete!")
}

func renderTemplate(w http.ResponseWriter, tmpl string, c *Content) {
	err := templates.ExecuteTemplate(w, tmpl+".html", c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

