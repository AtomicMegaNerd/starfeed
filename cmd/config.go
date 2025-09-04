package main

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"
)

const (
	ghTokenKey       = "STARFEED_GITHUB_API_TOKEN"
	freshRssUrlKey   = "STARFEED_FRESHRSS_URL"
	freshRssUserKey  = "STARFEED_FRESHRSS_USER"
	freshRssTokenKey = "STARFEED_FRESHRSS_API_TOKEN"
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey   = "STARFEED_HTTP_TIMEOUT"
)

type Config struct {
	GithubToken   string
	FreshRssUrl   string
	FreshRssUser  string
	FreshRssToken string
	DebugMode     bool
	SingleRunMode bool
	HttpTimeout   time.Duration
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

	// Parse optional HTTP timeout, default to 10 seconds
	httpTimeout := 10 * time.Second
	if timeoutStr := os.Getenv(httpTimeoutKey); timeoutStr != "" {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			httpTimeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	return &Config{
		GithubToken:   os.Getenv(ghTokenKey),
		FreshRssUrl:   os.Getenv(freshRssUrlKey),
		FreshRssUser:  os.Getenv(freshRssUserKey),
		FreshRssToken: os.Getenv(freshRssTokenKey),
		DebugMode:     os.Getenv(debugModeKey) == "true",
		SingleRunMode: os.Getenv(singleRunModeKey) == "true",
		HttpTimeout:   httpTimeout,
	}, nil
}
