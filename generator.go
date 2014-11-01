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
	posts         []*Post
	lastGenerated time.Time
}

type Post struct {
	Title        string
	Body         []byte
	Time         time.Time
	IsDraft      bool
	LastModified time.Time
}

// Rawr, a generator factory!
// TODO: Might want to remove this, for smaller API
func NewGenerator(dir string, lastGenerated time.Time) (*Generator, error) {
	posts, err := loadPostsFromDir(dir)
	if err != nil {
		return nil, err
	}

	g := &Generator{}
	g.lastGenerated = lastGenerated
	g.posts = posts

	return g, nil
}

func NewGeneratorWithPosts(ps []*Post, lastGenerated time.Time) *Generator {
	g := &Generator{}
	g.posts = ps
	g.lastGenerated = lastGenerated
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
		filepath := strings.Replace(post.Title, " ", "-", -1)
		filepath = strings.ToLower(filepath)
		filepath = path.Join(outDir, filepath)

		_, err := os.Stat(filepath)
		if post.IsDraft {
			if err == nil {
				os.RemoveAll(filepath)
			}
			continue
		}

		if err != nil && os.IsNotExist(err) {
			dirErr := os.Mkdir(filepath, 0776)
			if dirErr != nil {
				return dirErr
			}
			// There is no error, and the folder already exists
		} else if g.lastGenerated.Before(post.LastModified) || g.lastGenerated.Equal(post.LastModified) {
			// Delete index.html if it currently exists
			// TODO: Why do I need to delete?
			remPath := path.Join(filepath, "index.html")
			remErr := os.Remove(remPath)
			if remErr != nil {
				return remErr
			}
		}

		if g.lastGenerated.Before(post.LastModified) || g.lastGenerated.Equal(post.LastModified) {
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

	if len(filenameParts) < 3 {
		return nil, fmt.Errorf("%s has the wrong format!", name)
	}

	underscore := filenameParts[1]
	if underscore == "" {
		p.IsDraft = false
	} else {
		p.IsDraft = true
	}

	t, _ := time.Parse(layout, filenameParts[2])
	p.Time = t
	p.LastModified = fi.ModTime()

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
