package goblawg

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/feeds"
)

type Blog struct {
	Name         string
	Link         string
	Description  string
	Author       string
	Email        string
	Posts        []*Post
	InDir        string
	OutDir       string
	LastModified time.Time
}

func NewBlog(settingsJSON string) (*Blog, error) {
	dec := json.NewDecoder(strings.NewReader(settingsJSON))
	var b *Blog

	err := dec.Decode(&b)
	if err != nil && err != io.EOF {
		return nil, err
	}

	b.Posts, err = loadPostsFromDir(b.InDir)
	if err != nil {
		return nil, err
	}

	type timeDecode struct {
		LastGen string
	}
	var c timeDecode
	err = json.Unmarshal([]byte(settingsJSON), &c)
	if err != nil {
		return nil, err
	}

	tts, err := time.Parse(layout, c.LastGen)
	if err != nil {
		return nil, err
	}
	b.LastModified = tts

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

type ByTime []*Post

func (t ByTime) Len() int           { return len(t) }
func (t ByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTime) Less(i, j int) bool { return t[i].Time.Before(t[j].Time) }

// Return all published posts, sorted in reverse chronological order
func (b *Blog) GetPublishedPosts() []*Post {
	ps := []*Post{}
	for _, p := range b.Posts {
		if !p.IsDraft {
			ps = append(ps, p)
		}
	}
	sort.Sort(sort.Reverse(ByTime(ps)))
	return ps
}

// Generate the entire blog
// TODO: Copy over non-blog components
func (b *Blog) GenerateSite() error {
	g := NewGeneratorWithPosts(b.Posts, b.LastModified)

	err := g.GeneratePostsHTML(b.OutDir, "")
	if err != nil {
		return err
	}

	err = b.generateRSS()
	if err != nil {
		return err
	}

	b.LastModified = time.Now()

	return nil
}

// Generate the RSS feed
func (b *Blog) generateRSS() error {
	feed := &feeds.Feed{
		Title:       b.Name,
		Link:        &feeds.Link{Href: b.Link},
		Description: b.Description,
		Author:      &feeds.Author{b.Author, b.Email},
		Created:     time.Now(),
	}

	feed.Items = []*feeds.Item{}
	for _, p := range b.GetPublishedPosts() {
		desc := string(p.Body)
		if len(desc) > 120 {
			desc = desc[:120] + "..."
		}
		f := &feeds.Item{
			Title:       p.Title,
			Link:        &feeds.Link{Href: "None"}, // TODO
			Description: desc,
			Created:     p.Time,
		}
		feed.Items = append(feed.Items, f)
	}

	rss, err := feed.ToRss()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path.Join(b.OutDir, "feed.rss"), []byte(rss), 0776)
	if err != nil {
		return err
	}

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
