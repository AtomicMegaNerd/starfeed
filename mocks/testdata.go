package mocks

// This package contains mock data that we can re-use over and over again in our tests.
const (
	GitHubType    = "github"
	GitHubName    = "mygithub"
	GitHubURL     = "https://github.com"
	GitHubToken   = "ghp_abcdefghijklmnopqrstuvwxyz"
	ForgejoType   = "forgejo"
	ForgejoName   = "codeberg"
	ForgejoURL    = "https://codeberg.org"
	ForgejoToken  = "cb_abcdefghijklmnopqrstuvwxyz"
	FreshRSSType  = "freshrss"
	FreshRSSURL   = "http://freshrss.example.com"
	FreshRSSUser  = "testuser"
	FreshRSSToken = "freshrss_token456"
	// Use this whenever we need an invalid string
	Invalid      = "invalid"
	ZeroTimeout  = "0"
	ValidTimeout = "30"
	TrueBool     = "true"
)
