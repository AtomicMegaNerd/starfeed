package githost

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type GitHost interface {
	GetStarredRepos(context.Context) (map[string]Repo, error)
	IsReleaseFeed(string) bool
}

type Repo interface {
	Name() string
	FeedURL() string
}

// This is the config for each Git Host that we read from the environment. We read it from an
// environment variable called `STARFEED_GIT_HOST_ix` where ix is a number from 0..n.
// The format of the CSV is as follows:
// type,name,url,token
type GitHostConfig struct {
	Type    string `validate:"required,oneof=github forgejo"`
	Name    string `validate:"required,min=3"`
	BaseURL string `validate:"required,url"`
	Token   string `validate:"required,min=24"`
}

func ParseGitHostConfigFromCsv(csvLine string) (*GitHostConfig, error) {
	validate := validator.New()
	parts := strings.SplitN(csvLine, ",", 4)
	if len(parts) != 4 {
		return nil, errors.New("GitHostconfig invalid")
	}

	gitHost := &GitHostConfig{
		Type:    strings.TrimSpace(parts[0]),
		Name:    strings.TrimSpace(parts[1]),
		BaseURL: strings.TrimSpace(parts[2]),
		Token:   strings.TrimSpace(parts[3]),
	}

	return gitHost, validate.Struct(gitHost)
}

// This object represents a GitHub repo that is starred and that we want to
// get the Atom feed for.
type BaseRepo struct {
	RepoName string `json:"name"`
	RepoURL  string `json:"html_url"`
}

func (gr *BaseRepo) Name() string {
	return gr.RepoName
}

func (gr *BaseRepo) FeedURL() string {
	return fmt.Sprintf("%s/releases.atom", gr.RepoURL)
}

func (gr *BaseRepo) String() string {
	return fmt.Sprintf("Name: %s, Release Feed: %s", gr.Name(), gr.FeedURL())
}

// This is the response we get from GitHub
type GitHostResponse struct {
	Data     []byte
	NextPage string
}
