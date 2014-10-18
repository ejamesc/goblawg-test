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

var helloWorld = []byte("Hello world, this is my first post")

const layout = "2-Jan-2006-15-04-05"

func TestLoadPost(t *testing.T) {
	path := path.Join(os.TempDir(), "2-oct-2014-15-04-05-it-was-a-riot.md")
	ioutil.WriteFile(path, helloWorld, 0600)

	fi, _ := os.Stat(path)
	p, err := goblawg.LoadPost(path, fi)

	ok(t, err)
	time, err := time.Parse(layout, "2-Oct-2014-15-04-05")
	if err != nil {
		fmt.Printf("Crap, an errorL %s", err)
	}
	expected := &goblawg.Post{"It Was A Riot", helloWorld, time}
	equals(t, expected, p)

	// Teardown
	os.Remove(path)
}
