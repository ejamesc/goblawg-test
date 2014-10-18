package goblawg_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ejamesc/goblawg"
)

var bodyBytes = []byte("Hello world, this is my first post")

const layout = "2-Jan-2006-15-04-05"
const layout2 = "2 Jan 2006, 15:04:05"

// Ensure that LoadPost will throw an error on a non .md, .markdown or .txt file
func TestLoadPost_BadFile(t *testing.T) {
	path, fi := setup("", "superbadfilename")
	_, err := goblawg.LoadPost(path, fi)

	assert(t, err != nil, "err: %s", err)

	teardown(path)
}

// Ensure that LoadPost actually returns a well-formed, filled Post
func TestLoadPost(t *testing.T) {
	path, fi := setup("", "")

	p, err := goblawg.LoadPost(path, fi)
	ok(t, err)

	time, _ := time.Parse(layout, "2-Oct-2014-15-04-06")

	expected := &goblawg.Post{"It Was A Riot", bodyBytes, time}
	equals(t, expected, p)
	equals(t, "2 Oct 2014, 15:04:06", p.Time.Format(layout2))

	teardown(path)
}

type TestFileItem struct {
	Path     string
	FileInfo os.FileInfo
}

// Test Generator reads from the filesystem the right posts
func TestNewGenerator(t *testing.T) {
	// Setup
	dir := os.TempDir()
	filenames := []string{"12-Dec-2013-23-03-07-fade-away-love.markdown", "15-Aug-2014-09-08-06-it-was-a-riot.md", ""}

	// Setup bad file to show that NewGenerator ignores it
	badFilePath := path.Join(dir, "badfilename")
	ioutil.WriteFile(badFilePath, bodyBytes, 0600)

	var files []*TestFileItem
	for _, fname := range filenames {
		p, fi := setup(dir, fname)
		files = append(files, &TestFileItem{p, fi})
	}

	var posts []*goblawg.Post
	for _, tfi := range files {
		p, _ := goblawg.LoadPost(tfi.Path, tfi.FileInfo)
		posts = append(posts, p)
	}

	g, err := goblawg.NewGenerator(dir)

	ok(t, err)
	equals(t, posts, g.Posts)

	// Teardown
	for _, tfi := range files {
		teardown(tfi.Path)
	}
}

// Helpers
// Create files necessary for testing
func setup(pathname, filename string) (string, os.FileInfo) {
	if filename == "" {
		filename = "2-Oct-2014-15-04-06-it-was-a-riot.md"
	}

	resPath := ""
	if pathname == "" {
		resPath = path.Join(os.TempDir(), filename)
	} else {
		resPath = path.Join(pathname, filename)
	}
	ioutil.WriteFile(resPath, bodyBytes, 0600)

	fi, _ := os.Stat(resPath)
	return resPath, fi
}

func teardown(path string) {
	os.Remove(path)
}
