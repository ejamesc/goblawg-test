package goblawg

import (
	"io/ioutil"
	"os"
	"path"
	"time"
)

type Generator struct {
	Posts []Post
}

type Post struct {
	Title string
	Body  []byte
	Time  time.Time
}

func (g *Generator) LoadPosts() error {
	listFileInfo, err := ioutil.ReadDir("content/")
	if err != nil {
		return err
	}

	var markdownFileList []os.FileInfo
	for _, entry := range listFileInfo {
		if isMarkdownFile(entry.Name()) {
			markdownFileList = append(markdownFileList, entry)
		}
	}

	g.Posts = make([]Post, len(markdownFileList))
	for i, entry := range listFileInfo {
		p := Post{}
		p.LoadPost(entry)
		g.Posts[i] = p
	}

	return nil
}

func (p *Post) LoadPost(f os.FileInfo) {
	name := f.Name()
	p.Title = name
	// load body
}

func isMarkdownFile(n string) bool {
	ext := path.Ext(n)
	if ext == "md" || ext == "markdown" || ext == "txt" {
		return true
	} else {
		return false
	}
}
