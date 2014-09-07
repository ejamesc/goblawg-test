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

func (p *Post) save() error {
  filename := p.Title + ".md"
  return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPost(title string) (*Post, error) {
  filename := title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }

  htmlBody := blackfriday.MarkdownCommon(body)
  return &Post{Title: title, Body: htmlBody}, nil
}

var templates = template.Must(template.ParseFiles(
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
// List out all posts
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

func renderTemplate(w http.ResponseWriter, tmpl string, p *Post) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
