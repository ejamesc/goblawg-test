package main

import (
  "io/ioutil"
  "net/http"
  "html/template"
  
  "github.com/gorilla/mux"
  "github.com/russross/blackfriday"
)

type Post struct {
  Title string
  Body []byte
}

type Content struct {
  Title template.HTML
  Body template.HTML
}

func (p *Post) save() error {
  filename := p.Title + ".md"
  return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPost(title string) (*Content, error) {
  filename := title + ".md"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }

  htmlBody := blackfriday.MarkdownCommon(body)
  return &Content{Title: template.HTML(title), Body: template.HTML(htmlBody)}, nil
}

// TODO: better way to do this?  
var templates = template.Must(template.ParseFiles(
    "templates/partials.html",
    "templates/admin.html",
    "templates/view.html"))

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/admin/", adminHandler);
  r.HandleFunc("/view/{title}", viewHandler);

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
  if (err != nil) {
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
