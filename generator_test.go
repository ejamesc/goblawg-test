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

// Ensure that NewPostFromFile will throw an error on a non .md, .markdown or .txt file
func TestNewPostFromFile_BadFile(t *testing.T) {
	path, fi := setup("", "superbadfilename")
	_, err := goblawg.NewPostFromFile(path, fi)
	defer teardown(path)

	assert(t, err != nil, "err: %s", err)
}

func TestNewPostFromFile_Draft(t *testing.T) {
	path, fi := setup("", "_21-Oct-2013-14-06-10-the-shining.md")
	p, err := goblawg.NewPostFromFile(path, fi)
	defer teardown(path)

	ok(t, err)
	assert(t, p.IsDraft == true, "Post is not a draft")
}

// Ensure that NewPostFromFile actually returns a well-formed, filled Post
func TestNewPostFromFile(t *testing.T) {
	path, fi := setup("", "")

	p, err := goblawg.NewPostFromFile(path, fi)
	defer teardown(path)
	ok(t, err)

	tts, _ := time.Parse(layout, "2-Oct-2014-15-04-06")

	expected := &goblawg.Post{"It Was A Riot", bodyBytes, tts, false, fi.ModTime()}
	equals(t, expected, p)
	equals(t, "2 Oct 2014, 15:04:06", p.Time.Format(layout2))
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

	lastGenerated := time.Now().Add(-20 * time.Minute)
	g, err := goblawg.NewGenerator(dir, lastGenerated)

	// Teardown
	for _, tfi := range files {
		defer teardown(tfi.Path)
	}
	defer teardown(badFilePath)

	ok(t, err)
	assert(t, len(g.GetPosts()) == 4, "Expected 4 posts, instead got %v", len(g.GetPosts()))
	equals(t, posts, g.GetPosts())
}

var timeNow = time.Now()
var timeBefore = time.Now().Add(-20 * time.Minute)
var timeWayBefore = time.Now().Add(-40 * time.Minute)

var postFixtures = []*goblawg.Post{
	&goblawg.Post{"It Was A Riot", bodyBytes, time.Now(), false, timeNow},
	&goblawg.Post{"The World Tree", bodyBytes, time.Now(), false, timeNow},
	&goblawg.Post{"Fade Away Love", bodyBytes, time.Now(), false, timeWayBefore},
	&goblawg.Post{"Blah blah test", bodyBytes, time.Now(), true, timeNow},
}

// Test that we can create a Generator with a given list of posts
func TestNewGeneratorWithPosts(t *testing.T) {
	g := goblawg.NewGeneratorWithPosts(postFixtures, time.Now())

	equals(t, postFixtures, g.GetPosts())
}

// Test the ability to generate HTML from our posts
func TestGenerator_GeneratePostsHTML(t *testing.T) {
	g := goblawg.NewGeneratorWithPosts(postFixtures, time.Time{})

	dir := os.TempDir()
	err := g.GeneratePostsHTML(dir, "")

	ok(t, err)

	// Verify that the directory has posts generated to it
	fileInfoList, _ := ioutil.ReadDir(dir)
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
	// Teardown
	defer generator_teardown(dir, directories)

	// We expect the generate function to create 3 folders
	equals(t, len(dirNames), len(directories))
	// We expect the draft to not be created
	assert(t, draftExists == false, "Draft shouldn't exist.")
}

func TestGenerator_GeneratePostsHTMLAfterDateModified(t *testing.T) {
	// Setup
	dir := os.TempDir()
	os.Mkdir(path.Join(dir, "fade-away-love"), 0776)
	g := goblawg.NewGeneratorWithPosts(postFixtures, timeBefore)

	err := g.GeneratePostsHTML(dir, "")

	fileInfoList, _ := ioutil.ReadDir(dir)
	dirNames := []string{"it-was-a-riot", "the-world-tree", "fade-away-love"}
	var directories []os.FileInfo
	for _, fi := range fileInfoList {
		for _, name := range dirNames {
			if fi.Name() == name {
				directories = append(directories, fi)
			}
		}
	}
	// Teardown
	defer generator_teardown(dir, directories)

	ok(t, err)
	assert(t, len(directories) == 3, "Expected there to be 3 directories, got %v", len(directories))
	// Expect that index.html has not been created, because that would mean
	// the 'fade-away-love' folder wasn't touched.
	_, fileExistsErr := os.Stat(path.Join(dir, "fade-away-love", "index.html"))
	assert(t, fileExistsErr != nil, "Generator generates fade-away-love/index.html, when it should not")
}

// Test generating a post with a folder already created
func TestGenerator_GeneratePostsHTMLWithFolderCreated(t *testing.T) {
	g := goblawg.NewGeneratorWithPosts(postFixtures, time.Time{})

	dir := os.TempDir()

	folderName := path.Join(dir, "it-was-a-riot")
	os.Mkdir(folderName, 0776)
	ioutil.WriteFile(path.Join(folderName, "index.html"), bodyBytes, 0776)

	// Teardown
	defer teardownFolders(dir)

	err := g.GeneratePostsHTML(dir, "")
	ok(t, err)

}

func teardownFolders(dir string) {
	dirNames := []string{"it-was-a-riot", "the-world-tree", "fade-away-love"}
	fileInfoList, _ := ioutil.ReadDir(dir)
	for _, f := range fileInfoList {
		if f.IsDir() {
			for _, name := range dirNames {
				if name == f.Name() {
					tmpPath := path.Join(dir, name)
					os.RemoveAll(tmpPath)
				}
			}
		}
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

func generator_teardown(dir string, generatedDirs []os.FileInfo) {
	for _, tmpDir := range generatedDirs {
		tempPath := path.Join(dir, tmpDir.Name())
		os.RemoveAll(tempPath)
	}
}
