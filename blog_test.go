package goblawg_test

import (
	"fmt"
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
	g := goblawg.NewGeneratorWithPosts(postList, testTime)

	b, err := goblawg.NewBlog(settingsJSON)

	// Teardown
	defer teardown(path)

	ok(t, err)

	equals(t, b.Name, "My First Blog")
	equals(t, b.OutDir, dir)
	equals(t, b.InDir, dir)
	equals(t, b.Posts, postList)
	equals(t, b.Generator, g)
}
