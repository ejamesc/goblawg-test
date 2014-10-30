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
	Name         string
	Posts        []*Post
	InDir        string
	OutDir       string
	LastModified time.Time
}

type config struct {
	Name    string
	InDir   string
	OutDir  string
	LastGen string
}

func NewBlog(settingsJSON string) (*Blog, error) {
	dec := json.NewDecoder(strings.NewReader(settingsJSON))
	var c config

	err := dec.Decode(&c)
	if err != nil && err != io.EOF {
		return nil, err
	}

	posts, err := loadPostsFromDir(c.InDir)
	if err != nil {
		return nil, err
	}

	tts, err := time.Parse(layout, c.LastGen)
	if err != nil {
		return nil, err
	}

	b := &Blog{}
	b.Name = c.Name
	b.Posts = posts
	b.InDir = c.InDir
	b.LastModified = tts
	b.OutDir = c.OutDir

	return b, nil
}

// Save a blog post and write to disk
func (b *Blog) SavePost(post *Post) error {
	title := strings.Replace(post.Title, " ", "-", -1)
	title = strings.ToLower(title)
	timeString := post.Time.Format(layout)
	filename := timeString + "-" + title + ".md"

	if post.IsDraft {
		filename = "_" + filename
	}

	filepath := path.Join(b.OutDir, filename)
	err := ioutil.WriteFile(filepath, post.Body, 0776)
	if err != nil {
		return err
	}

	b.Posts = append(b.Posts, post)
	return nil
}

// Generate the entire blog
// TODO: Copy over non-blog components
func (b *Blog) GenerateHTML() error {
	g := NewGeneratorWithPosts(b.Posts, b.LastModified)

	err := g.GeneratePostsHTML(b.OutDir, "")
	if err != nil {
		return err
	}

	b.LastModified = time.Now()

	return nil
}

// Helpers
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
