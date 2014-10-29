package goblawg_test

import (
	"fmt"
	"io/ioutil"
	"os"
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
	settingsJSON := fmt.Sprintf(`{"Name": "My First Blog", "OutDir": "%s", "InDir": "%s", "LastGen": "12-Jan-2014-15-05-02"}`, dir, dir)

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
	path, fi := setup(dir, "")
	defer teardown(path)

	tts, _ := time.Parse(layout, "21-Oct-2013-14-06-10")

	post1, _ := goblawg.NewPostFromFile(path, fi)
	post2 := &goblawg.Post{"The Shining", bodyBytes, tts, true, time.Now()}

	postListBefore := []*goblawg.Post{post1}
	postListAfter := []*goblawg.Post{post1, post2}

	b := &goblawg.Blog{Posts: postListBefore, OutDir: dir}
	err := b.SavePost(post2)

	ok(t, err)
	equals(t, b.Posts, postListAfter)

	fileInfoList, _ := ioutil.ReadDir(dir)
	fileGenerated := false
	for _, fi := range fileInfoList {
		if fi.Name() == "_21-Oct-2013-14-06-10-the-shining.md" {
			fileGenerated = true
		}
	}

	assert(t, fileGenerated == true, "Post not generated")
}
