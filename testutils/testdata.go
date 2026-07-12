package testutils

// This package contains mock data that we can re-use over and over again in our tests.
const (
	GitHubName  = "mygithub"
	GitHubURL   = "https://github.com"
	GitHubToken = "ghp_abcdefghijklmnopqrstuvwxyz"
	// Forgejo/Codeberg
	CodebergName  = "codeberg"
	CodebergURL   = "https://codeberg.org"
	CodebergToken = "cb_abcdefghijklmnopqrstuvwxyz"
	// FreshRSS
	FreshRSSName  = "freshrss"
	FreshRSSURL   = "http://freshrss.example.com"
	FreshRSSUser  = "testuser@email.com"
	FreshRSSToken = "freshrss_token456"
	// Use this whenever we need an invalid string
	Invalid      = "invalid"
	ZeroTimeout  = "0"
	ValidTimeout = "30"
	TrueBool     = "true"
)
