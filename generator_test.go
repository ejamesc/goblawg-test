package goblawg_test

import (
	"fmt"
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

// Ensure that NewPostFromFile will throw an error on a non .md, .markdown or .txt file
func TestNewPostFromFile_BadFile(t *testing.T) {
	path, fi := setup("", "superbadfilename")
	_, err := goblawg.NewPostFromFile(path, fi)

	assert(t, err != nil, "err: %s", err)

	teardown(path)
}

func TestNewPostFromFile_Draft(t *testing.T) {
	path, fi := setup("", "_21-Oct-2013-14-06-10-the-shining.md")
	p, err := goblawg.NewPostFromFile(path, fi)

	ok(t, err)
	assert(t, p.IsDraft == true, "")

	teardown(path)
}

// Ensure that NewPostFromFile actually returns a well-formed, filled Post
func TestNewPostFromFile(t *testing.T) {
	path, fi := setup("", "")

	p, err := goblawg.NewPostFromFile(path, fi)
	ok(t, err)

	time, _ := time.Parse(layout, "2-Oct-2014-15-04-06")

	expected := &goblawg.Post{"It Was A Riot", bodyBytes, time, false}
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
	filenames := []string{"12-Dec-2013-23-03-07-fade-away-love.markdown", "15-Aug-2014-09-08-06-the-world-tree.md", "", "_13-Oct-2014-23-07-08-this-is-draft.md"}

	var files []*TestFileItem
	for _, fname := range filenames {
		p, fi := setup(dir, fname)
		files = append(files, &TestFileItem{p, fi})
	}

	var posts []*goblawg.Post
	for _, tfi := range files {
		p, _ := goblawg.NewPostFromFile(tfi.Path, tfi.FileInfo)
		posts = append(posts, p)
	}

	// Setup bad file to show that NewGenerator ignores it
	badFilePath := path.Join(dir, "badfilename")
	ioutil.WriteFile(badFilePath, bodyBytes, 0600)

	g, err := goblawg.NewGenerator(dir)

	ok(t, err)
	assert(t, len(g.GetPosts()) == 4, "")
	equals(t, posts, g.GetPosts())

	// Teardown
	for _, tfi := range files {
		teardown(tfi.Path)
	}
	teardown(badFilePath)
}

var postFixtures = []*goblawg.Post{
	&goblawg.Post{"It Was A Riot", bodyBytes, time.Now(), false},
	&goblawg.Post{"The World Tree", bodyBytes, time.Now(), false},
	&goblawg.Post{"Fade Away Love", bodyBytes, time.Now(), false},
	&goblawg.Post{"Blah blah test", bodyBytes, time.Now(), true},
}

func TestNewGeneratorWithPosts(t *testing.T) {
	g := goblawg.NewGeneratorWithPosts(postFixtures)

	equals(t, postFixtures, g.GetPosts())
}

func TestGenerator_GeneratePostsHTML(t *testing.T) {
	g := goblawg.NewGeneratorWithPosts(postFixtures)

	dir := os.TempDir()
	err := g.GeneratePostsHTML(dir, "")

	ok(t, err)

	// Verify that the directory has posts generated to it
	fileInfoList, rErr := ioutil.ReadDir(dir)
	if rErr != nil {
		fmt.Println("ReadDir error, dammit: %s", err)
	}

	dirNames := []string{"it-was-a-riot", "the-world-tree", "fade-away-love"}
	draftExists := false
	var directories []os.FileInfo

	for _, f := range fileInfoList {
		if f.IsDir() {
			if f.Name() == "blah-blah-test" {
				draftExists = true
			}
			for _, name := range dirNames {
				if name == f.Name() {
					directories = append(directories, f)
				}
			}
		}
	}

	// We expect the generate function to create the 3 folders
	equals(t, len(dirNames), len(directories))
	// We expect the draft to not be created
	assert(t, draftExists == false, "")

	// Teardown
	for _, tmpDir := range directories {
		tempPath := path.Join(dir, tmpDir.Name())
		os.RemoveAll(tempPath)
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
