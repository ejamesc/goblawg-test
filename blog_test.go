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

// Test NewBlog constructs and returns a Blog struct correctly
func TestNewBlog(t *testing.T) {
	// Setup
	dir := os.TempDir()
	path, fi := setup(dir, "")
	post, _ := goblawg.NewPostFromFile(path, fi)
	settingsJSON := fmt.Sprintf(
		`{"Name": "My First Blog", 
	"OutDir": "%s", 
	"InDir": "%s", 
	"LastGen": "12-Jan-2014-15-05-02",
	"Link": "http://elijames.org",
	"Description": "Test Blog",
	"Author": "Eli James",
	"Email": "bob@test.com"}`, dir, dir)

	postList := []*goblawg.Post{post}
	testTime, _ := time.Parse(layout, "12-Jan-2014-15-05-02")

	b, err := goblawg.NewBlog(settingsJSON)

	// Teardown
	defer teardown(path)

	ok(t, err)

	equals(t, b.Name, "My First Blog")
	equals(t, b.OutDir, dir)
	equals(t, b.InDir, dir)
	equals(t, b.Posts, postList)
	equals(t, b.LastModified, testTime)
}

// Test saving posts
func TestSavePost(t *testing.T) {
	//Setup
	dir := os.TempDir()
	testPath, fi := setup(dir, "")
	defer teardown(testPath)

	tts, _ := time.Parse(layout, "21-Oct-2013-14-06-10")
	currTime := time.Now()

	post1, _ := goblawg.NewPostFromFile(testPath, fi)
	post2 := &goblawg.Post{"The Shining", bodyBytes, tts, true, currTime}

	postListBefore := []*goblawg.Post{post1}
	postListAfter := []*goblawg.Post{post1, post2}

	b := &goblawg.Blog{Posts: postListBefore, OutDir: dir}
	err := b.SavePost(post2)

	ok(t, err)
	equals(t, b.Posts, postListAfter)

	fileInfoList, _ := ioutil.ReadDir(dir)
	fileGenerated := false
	var postUnderTest *goblawg.Post
	for _, fi := range fileInfoList {
		if fi.Name() == "_21-Oct-2013-14-06-10-the-shining.md" {
			fileGenerated = true
			filePath := path.Join(dir, fi.Name())
			postUnderTest, _ = goblawg.NewPostFromFile(filePath, fi)
			defer teardown(filePath)
		}
	}

	assert(t, fileGenerated == true, "Post not generated")
	// Stub out the LastModified
	postUnderTest.LastModified = currTime
	equals(t, post2, postUnderTest)
}

// Test Generate HTML
func TestGenerateSite(t *testing.T) {
	// Setup
	dir := os.TempDir()
	post := &goblawg.Post{"The Shining", bodyBytes, time.Now(), false, time.Now()}

	b := &goblawg.Blog{Posts: []*goblawg.Post{post}, LastModified: time.Time{}, OutDir: dir}
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
