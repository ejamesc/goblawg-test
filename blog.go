package goblawg

type Blog struct {
	Generator Generator
	Name      string
	Posts     []*Post
	InDir     string
	OutDir    string
}

func NewBlog(settings string) *Blog {
	//dec := json.NewDecoder(strings.NewReader(settings))

	p := &Blog{}
	return p
}
