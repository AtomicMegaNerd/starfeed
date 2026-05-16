package forgejo

import "fmt"

// This object represents a Forgejo repo that is starred and that we want to
// get the Atom feed for.
type ForgejoRepo struct {
	RepoName string `json:"name"`
	HTMLURL  string `json:"html_url"`
}

func (gr *ForgejoRepo) Name() string {
	return gr.RepoName
}

func (gr *ForgejoRepo) FeedURL() string {
	return fmt.Sprintf("%s/releases.atom", gr.HTMLURL)
}

func (gr *ForgejoRepo) String() string {
	return fmt.Sprintf("Name: %s, Release Feed: %s", gr.Name(), gr.FeedURL())
}
