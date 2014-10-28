package goblawg

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

type Blog struct {
	Generator *Generator
	Name      string
	Posts     []*Post
	InDir     string
	OutDir    string
}

type config struct {
	Name    string
	InDir   string
	OutDir  string
	LastGen string
}

func NewBlog(settings string) (*Blog, error) {
	dec := json.NewDecoder(strings.NewReader(settings))
	var c config

	err := dec.Decode(&c)
	if err != nil && err != io.EOF {
		return nil, err
	}

	posts, err := loadPostsFromDir(c.InDir)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(layout, c.LastGen)
	if err != nil {
		return nil, err
	}

	g := NewGeneratorWithPosts(posts, t)

	p := &Blog{}
	p.Name = c.Name
	p.Posts = posts
	p.Generator = g
	p.InDir = c.InDir
	p.OutDir = c.OutDir

	return p, nil
}

func loadPostsFromDir(dir string) ([]*Post, error) {
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

	posts := make([]*Post, len(markdownFileList))

	for i, entry := range markdownFileList {
		fpath := path.Join(dir, entry.Name())

		p, err := NewPostFromFile(fpath, entry)
		if err != nil {
			return nil, err
		}
		posts[i] = p
	}

	return posts, nil
}
