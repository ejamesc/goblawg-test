package goblawg

import (
	"io/ioutil"
	"os"
)

type Generator struct {
	Posts []Post
}

type Post struct {
	Title string
	Body  []byte
}

func (g *Generator) LoadPosts() error {
	listFileInfo, err := ioutil.ReadDir("content/")
	if err != nil {
		return err
	}

	g.Posts = make([]Post, len(listFileInfo))
	for i, entry := range listFileInfo {
		p := make(Post{})
		g.Posts[i] = p.LoadPost(entry)
	}

	return nil
}

func (p *Post) LoadPost(f os.FileInfo) {
	name := f.Name
	p.Title = name
	// load body
}
