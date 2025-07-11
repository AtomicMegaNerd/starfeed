package main

import (
	"errors"
	"log/slog"
	"os"
	"time"
)

const (
	ghTokenKey       = "STARFEED_GITHUB_API_TOKEN"
	freshRssUrlKey   = "STARFEED_FRESHRSS_URL"
	freshRssUserKey  = "STARFEED_FRESHRSS_USER"
	freshRssTokenKey = "STARFEED_FRESHRSS_API_TOKEN"
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"

	httpTimeoutInSeconds = 10
)

type Config struct {
	GithubToken   string
	FreshRssUrl   string
	FreshRssUser  string
	FreshRssToken string
	DebugMode     bool
	HttpTimeout   time.Duration
	SingleRunMode bool
}

func NewConfig() (*Config, error) {
	// Check for required environment variables
	if os.Getenv(ghTokenKey) == "" ||
		os.Getenv(freshRssUrlKey) == "" ||
		os.Getenv(freshRssUserKey) == "" ||
		os.Getenv(freshRssTokenKey) == "" {
		slog.Error("Missing required environment variables")
		return nil, errors.New("missing required environment variables")
	}

	return &Config{
		GithubToken:   os.Getenv(ghTokenKey),
		FreshRssUrl:   os.Getenv(freshRssUrlKey),
		FreshRssUser:  os.Getenv(freshRssUserKey),
		FreshRssToken: os.Getenv(freshRssTokenKey),
		DebugMode:     os.Getenv(debugModeKey) == "true",
		HttpTimeout:   httpTimeoutInSeconds,
		SingleRunMode: os.Getenv(singleRunModeKey) == "true",
	}, nil
}
