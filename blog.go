package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
  "strings"

	"github.com/gorilla/mux"
	"github.com/russross/blackfriday"
)

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

func convertToFilename(urlPath string) string {
  res := strings.Replace(urlPath, "-", "_", -1)
  res = "content/" + res + ".md"
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

  //w := 
  //templates.ExecuteTemplate(w, template+".html", content)
  return nil
}

func loadPost(title string) (*Content, error) {
  filename := convertToFilename(title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	htmlBody := blackfriday.MarkdownCommon(body)
  displayTitle := convertToDisplayTitle(title)
	return &Content{Title: template.HTML(displayTitle), Body: template.HTML(htmlBody)}, nil
}

// TODO: better way to do this?
var templates = template.Must(template.ParseFiles(
	"templates/partials.html",
	"templates/admin.html",
	"templates/view.html"))

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/admin/", adminHandler)
	r.HandleFunc("/view/{title}", viewHandler)

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

/* HANDLERS */
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

func renderTemplate(w http.ResponseWriter, tmpl string, c *Content) {
	err := templates.ExecuteTemplate(w, tmpl+".html", c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

