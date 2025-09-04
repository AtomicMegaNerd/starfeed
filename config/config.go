package config

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type EnvGetter interface {
	Getenv(key string) string
}

type OSEnvGetter struct{}

func (o OSEnvGetter) Getenv(key string) string {
	return os.Getenv(key)
}

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

func NewConfig(envGetter EnvGetter) (*Config, error) {
	// Check for required environment variables
	if envGetter.Getenv(ghTokenKey) == "" ||
		envGetter.Getenv(freshRssUrlKey) == "" ||
		envGetter.Getenv(freshRssUserKey) == "" ||
		envGetter.Getenv(freshRssTokenKey) == "" {
		slog.Error("Missing required environment variables")
		return nil, errors.New("missing required environment variables")
	}

	// Parse optional HTTP timeout, default to 10 seconds
	httpTimeout := 10 * time.Second
	if timeoutStr := envGetter.Getenv(httpTimeoutKey); timeoutStr != "" {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			httpTimeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	return &Config{
		GithubToken:   envGetter.Getenv(ghTokenKey),
		FreshRssUrl:   envGetter.Getenv(freshRssUrlKey),
		FreshRssUser:  envGetter.Getenv(freshRssUserKey),
		FreshRssToken: envGetter.Getenv(freshRssTokenKey),
		DebugMode:     envGetter.Getenv(debugModeKey) == "true",
		SingleRunMode: envGetter.Getenv(singleRunModeKey) == "true",
		HttpTimeout:   httpTimeout,
	}, nil
}
