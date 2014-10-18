package goblawg

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

type Generator struct {
	Posts []*Post
}

type Post struct {
	Title string
	Body  []byte
	Time  time.Time
}

var r = regexp.MustCompile(`(\d{1,2}-[a-z]{3}-\d{4}-\d{1,2}-\d{1,2}-\d{1,2})-(.)`)

const layout = "2-Jan-2006-15-04-05"

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
	g.Posts = make([]*Post, len(markdownFileList))
	for i, entry := range listFileInfo {
		path := path.Join(dir, entry.Name())
		// TODO: why do I need to have a reference?
		p, err := LoadPost(path, entry)
		if err != nil {
			return nil, err
		}

		g.Posts[i] = p
	}

	return g, nil
}

func LoadPost(path string, fi os.FileInfo) (*Post, error) {
	p := &Post{}

	name := fi.Name()
	// Assume filename is 2-jan-2006-15-04-05-it-was-a-riot.md
	m := r.FindStringSubmatch(name)
	t, _ := time.Parse(layout, m[1])
	p.Time = t

	filename := m[2]
	filename_parts := strings.Split(filename, ".")
	title := strings.Replace(filename_parts[0], "-", " ", -1)
	p.Title = strings.ToUpper(title)

	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p.Body = body

	return p, nil
}

func (p *Post) LoadPost(f os.FileInfo) {
	name := f.Name()
	p.Title = name
	// load body
}

// Helper Functions
func isMarkdownFile(n string) bool {
	ext := path.Ext(n)
	if ext == "md" || ext == "markdown" || ext == "txt" {
		return true
	} else {
		return false
	}
}
