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

var tmpdir = os.TempDir()
var settingsJSON = fmt.Sprintf(
	`{"Name": "My First Blog", 
	"OutDir": "%s", 
	"InDir": "%s", 
	"LastGen": "12-Jan-2014-15-05-02",
	"Link": "http://elijames.org",
	"Description": "Test Blog",
	"Author": "Eli James",
	"Email": "bob@test.com"}`, tmpdir, tmpdir)

// Test NewBlog constructs and returns a Blog struct correctly
func TestNewBlog(t *testing.T) {
	// Setup
	blogpath := path.Join(tmpdir, "posts")
	os.Mkdir(blogpath, 0775)
	path, fi := setup(blogpath, "")
	post, _ := goblawg.NewPostFromFile(path, fi)

	postList := []*goblawg.Post{post}
	testTime, _ := time.Parse(layout, "12-Jan-2014-15-05-02")

	b, err := goblawg.NewBlog(settingsJSON)

	// Teardown
	defer os.RemoveAll(blogpath)

	ok(t, err)

	equals(t, b.Name, "My First Blog")
	equals(t, b.OutDir, tmpdir)
	equals(t, b.InDir, tmpdir)
	equals(t, b.Posts, postList)
	equals(t, b.LastModified, testTime)
}

// Test saving posts
func TestSaveAndDeletePost(t *testing.T) {
	//Setup
	dir := os.TempDir()
	testPath, fi := setup(dir, "")
	defer teardown(testPath)

	tts, _ := time.Parse(layout, "21-Oct-2013-14-06-10")
	currTime := time.Now()

	post1, _ := goblawg.NewPostFromFile(testPath, fi)
	post2 := &goblawg.Post{"The Shining", []byte("Hello world, this is my first post"), "the-shining", tts, true, currTime}

	postListBefore := []*goblawg.Post{post1}
	postListAfter := []*goblawg.Post{post1, post2}

	b := &goblawg.Blog{Posts: postListBefore, InDir: dir}
	err := b.SavePost(post2)

	ok(t, err)
	equals(t, b.Posts, postListAfter)

	postPath := path.Join(dir, "posts")
	fileInfoList, _ := ioutil.ReadDir(postPath)
	fileGenerated := false
	var postUnderTest *goblawg.Post
	for _, fi := range fileInfoList {
		if fi.Name() == "_21-Oct-2013-14-06-10-the-shining.md" {
			fileGenerated = true
			filePath := path.Join(postPath, fi.Name())
			postUnderTest, _ = goblawg.NewPostFromFile(filePath, fi)
			defer teardown(filePath)
		}
	}

	assert(t, fileGenerated == true, "Post not generated")
	// Stub out the LastModified
	postUnderTest.LastModified = currTime
	equals(t, post2, postUnderTest)

	// Test the deletion
	err = b.DeletePost(post2)

	ok(t, err)
	assert(t, len(b.Posts) == 1, "Post not deleted from b.Posts")
	fList, _ := ioutil.ReadDir(postPath)
	fileDeleted := true
	for _, fi := range fList {
		if fi.Name() == "_21-Oct-2013-14-06-10-the-shining.md" {
			fileDeleted = false
		}
	}
	assert(t, fileDeleted == true, "Post not deleted from file system")

	// Test deletion error works
	err = b.DeletePost(post2)
	assert(t, err != nil, "Expecting error to be returned when deleting non-existent post")
}

// Test Generate HTML
func TestGenerateSite(t *testing.T) {
	// Setup
	dir := os.TempDir()
	post := &goblawg.Post{"The Shining", bodyBytes, "the-shining", time.Now(), false, time.Now()}

	b := &goblawg.Blog{Posts: []*goblawg.Post{post}, LastModified: time.Time{}, InDir: dir, OutDir: dir}
	err := b.GenerateSite()

	// Teardown
	generatedPath := path.Join(dir, "the-shining")
	defer os.RemoveAll(generatedPath)

	ok(t, err)
	assert(t, b.LastModified != time.Time{}, "Expected last modified timestamp to have been updated")

	_, err1 := os.Stat(generatedPath)
	_, err2 := os.Stat(path.Join(generatedPath, "index.html"))
	ok(t, err1)
	ok(t, err2)
}

// Test that GetPosts returns a reverse chronological list of posts
func TestGetPublishedPosts(t *testing.T) {
	postFixtures[1].Time = timeWayBefore
	postFixtures[2].Time = timeBefore

	b := &goblawg.Blog{Posts: postFixtures}
	posts := b.GetPublishedPosts()

	orderedPostList := []*goblawg.Post{postFixtures[0], postFixtures[2], postFixtures[1]}
	equals(t, posts, orderedPostList)
}

// Test the RSS and atom feeds are generated.
func TestGenerateRSS(t *testing.T) {
	os.Mkdir(path.Join(tmpdir, "posts"), 0775)

	b, _ := goblawg.NewBlog(settingsJSON)
	b.Posts = postFixtures
	// Test truncation of post description
	b.Posts[0].Body = []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")
	err := b.GenerateRSS()

	defer func() {
		teardown(path.Join(tmpdir, "feed.rss"))
		os.RemoveAll(path.Join(tmpdir, "blog"))
	}()

	ok(t, err)

	fiList, _ := ioutil.ReadDir(tmpdir)
	filteredList := filterDir(fiList, func(fi os.FileInfo) bool {
		if fi.Name() == "feed.rss" {
			return true
		}
		return false
	})

	assert(t, len(filteredList) == 1, "Boo, feed.rss wasn't generated")
	assert(t, filteredList[0].Size() > 0, "feed.rss appears to be empty!")
}

// Ensure that the rest of the site is generated perfectly
func testGeneratingSitePages(t *testing.T) {
	// Setup
	b, _ := goblawg.NewBlog(settingsJSON)
	aPath := path.Join(b.InDir, "about.html")
	ioutil.WriteFile(aPath, bodyBytes, 0775)

	err := b.GenerateSitePages()

	// Teardown
	defer func() {
		os.Remove(aPath)
		os.RemoveAll(path.Join(b.OutDir, "about"))
	}()

	ok(t, err)

	fi, err := os.Stat(path.Join(b.OutDir, "about"))
	assert(t, err == nil, "Error with generation of about/index.html: %s", err)
	assert(t, fi.IsDir() == true, "Expected about/ to be a folder, but got a file.")
}

func testGetPostByLink(t *testing.T) {
	b, _ := goblawg.NewBlog(settingsJSON)
	b.Posts = postFixtures

	postUnderTest := b.GetPostByLink("it-was-a-riot")
	assert(t, postUnderTest != nil, "Post not retrieved successfully")
	assert(t, postUnderTest.Title == "It Was A Riot", "Wrong post was retrieved")
}
