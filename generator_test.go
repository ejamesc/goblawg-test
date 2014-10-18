package goblawg_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ejamesc/goblawg"
)

var helloWorld = []byte("Hello world, this is my first post")

const layout = "2-Jan-2006-15-04-05"
const layout2 = "2 Jan 2006, 15:04:05"

func setup() (string, os.FileInfo) {
	path := path.Join(os.TempDir(), "2-Oct-2014-15-04-06-it-was-a-riot.md")
	ioutil.WriteFile(path, helloWorld, 0600)

	fi, _ := os.Stat(path)
	return path, fi
}

func teardown(path string) {
	os.Remove(path)
}

func TestLoadPost(t *testing.T) {
	path, fi := setup()

	p, err := goblawg.LoadPost(path, fi)
	ok(t, err)

	time, _ := time.Parse(layout, "2-Oct-2014-15-04-06")

	expected := &goblawg.Post{"It Was A Riot", helloWorld, time}
	equals(t, expected, p)
	equals(t, "2 Oct 2014, 15:04:06", p.Time.Format(layout2))

	teardown(path)
}
