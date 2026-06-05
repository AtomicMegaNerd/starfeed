package testutils

// This package contains mock data that we can re-use over and over again in our tests.
const (
	// Github
	GitHubHostType = "github"
	GitHubName     = "mygithub"
	GitHubURL      = "https://github.com"
	GitHubToken    = "ghp_abcdefghijklmnopqrstuvwxyz"
	GitHubAPIURL   = "https://api.github.com"
	// Forgejo/Codeberg
	ForgejoHostType = "forgejo"
	CodebergName    = "codeberg"
	CodebergURL     = "https://codeberg.org"
	CodebergAPIURL  = "https://api.codeberg.org"
	CodebergToken   = "cb_abcdefghijklmnopqrstuvwxyz"
	// FreshRSS
	FreshRSSType  = "freshrss"
	FreshRSSURL   = "http://freshrss.example.com"
	FreshRSSUser  = "testuser@email.com"
	FreshRSSToken = "freshrss_token456"
	// Use this whenever we need an invalid string
	Invalid      = "invalid"
	ZeroTimeout  = "0"
	ValidTimeout = "30"
	TrueBool     = "true"
)
