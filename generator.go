package goblawg

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"
	"time"
)

var r = regexp.MustCompile(`(_*)(\d{1,2}-[a-zA-Z]{3}-\d{4}-\d{1,2}-\d{1,2}-\d{1,2})-(.+)`)

const layout = "2-Jan-2006-15-04-05"

type Generator struct {
	posts []*Post
}

type Post struct {
	Title   string
	Body    []byte
	Time    time.Time
	IsDraft bool
}

// Rawr, a generator factory!
func NewGenerator(dir string) (*Generator, error) {
	listFileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var markdownFileList []os.FileInfo
	for _, entry := range listFileInfo {
		if isMarkdownFile(entry.Name()) {
			markdownFileList = append(markdownFileList, entry)
		}
	}

	g := &Generator{}
	g.posts = make([]*Post, len(markdownFileList))
	for i, entry := range markdownFileList {
		fpath := path.Join(dir, entry.Name())
		p, err := NewPostFromFile(fpath, entry)
		if err != nil {
			return nil, err
		}

		g.posts[i] = p
	}

	return g, nil
}

func NewGeneratorWithPosts(ps []*Post) *Generator {
	g := &Generator{ps}
	return g
}

// Return the array of posts
func (g *Generator) GetPosts() []*Post {
	return g.posts
}

// Generates just the HTML version of the posts
func (g *Generator) GeneratePostsHTML(outDir, templateLoc string) error {
	if templateLoc == "" {
		templateLoc = "templates/essay.html"
	}

	for _, post := range g.posts {
		if post.IsDraft {
			continue
		}
		filepath := strings.Replace(post.Title, " ", "-", -1)
		filepath = strings.ToLower(filepath)
		filepath = path.Join(outDir, filepath)

		_, err := os.Stat(filepath)
		if err != nil && os.IsNotExist(err) {
			dirErr := os.Mkdir(filepath, 0776)
			if dirErr != nil {
				return dirErr
			}
		} else {
			// Delete folder if it currently exists
			remErr := os.RemoveAll(filepath)
			if remErr != nil {
				return remErr
			}
		}
		filepath = path.Join(filepath, "index.html")

		file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0776)
		defer file.Close()
		if err != nil {
			return err
		}

		t, err := template.ParseFiles(templateLoc)
		if err != nil {
			return err
		}

		t.Execute(file, post)
	}

	return nil
}

// Create a new post from file
func NewPostFromFile(path string, fi os.FileInfo) (*Post, error) {
	if !isMarkdownFile(path) {
		return nil, fmt.Errorf("%s does not have a markdown or text file extension.", path)
	}

	p := &Post{}

	name := fi.Name()
	filenameParts := r.FindStringSubmatch(name)

	underscore := filenameParts[1]
	if underscore == "" {
		p.IsDraft = false
	} else {
		p.IsDraft = true
	}

	t, _ := time.Parse(layout, filenameParts[2])
	p.Time = t

	filename := filenameParts[3]
	filename_parts := strings.Split(filename, ".")
	title := strings.Replace(filename_parts[0], "-", " ", -1)
	p.Title = strings.Title(title)

	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p.Body = body

	return p, nil
}

// Helper Functions

func isMarkdownFile(n string) bool {
	ext := path.Ext(n)
	if ext == ".md" || ext == ".markdown" || ext == ".txt" {
		return true
	} else {
		return false
	}
}
