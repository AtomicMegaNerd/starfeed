package config

import (
	"errors"
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

func parseBoolEnv(envGetter EnvGetter, key string) (bool, error) {
	v := envGetter.Getenv(key)
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}

const (
	ghTokenKey              = "STARFEED_GITHUB_API_TOKEN"
	freshRSSURLKey          = "STARFEED_FRESHRSS_URL"
	freshRSSUserKey         = "STARFEED_FRESHRSS_USER"
	freshRSSTokenKey        = "STARFEED_FRESHRSS_API_TOKEN"
	debugModeKey            = "STARFEED_DEBUG_MODE"
	singleRunModeKey        = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey          = "STARFEED_HTTP_TIMEOUT"
	disableRepoFeedModeKey  = "STARFEED_DISABLE_REPO_FEED_MODE"
	disableIssueFeedModeKey = "STARFEED_DISABLE_ISSUE_FEED_MODE"
	disablePRFeedModeKey    = "STARFEED_DISABLE_PR_FEED_MODE"
)

type Config struct {
	GitHubToken          string
	FreshRSSURL          string
	FreshRSSUser         string
	FreshRSSToken        string
	DebugMode            bool
	SingleRunMode        bool
	DisableRepoFeedMode  bool
	DisableIssueFeedMode bool
	DisablePRFeedMode    bool
	HTTPTimeout          time.Duration
}

func NewConfig(envGetter EnvGetter) (*Config, error) {
	// Check for required environment variables
	if envGetter.Getenv(ghTokenKey) == "" ||
		envGetter.Getenv(freshRSSURLKey) == "" ||
		envGetter.Getenv(freshRSSUserKey) == "" ||
		envGetter.Getenv(freshRSSTokenKey) == "" {
		return nil, errors.New("missing required environment variables")
	}

	// Parse optional HTTP timeout, default to 10 seconds
	httpTimeout := 10 * time.Second
	if timeoutStr := envGetter.Getenv(httpTimeoutKey); timeoutStr != "" {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			httpTimeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	debugMode, err := parseBoolEnv(envGetter, debugModeKey)
	if err != nil {
		return nil, err
	}
	singleRunMode, err := parseBoolEnv(envGetter, singleRunModeKey)
	if err != nil {
		return nil, err
	}
	disableRepoFeedMode, err := parseBoolEnv(envGetter, disableRepoFeedModeKey)
	if err != nil {
		return nil, err
	}
	disableIssueFeedMode, err := parseBoolEnv(envGetter, disableIssueFeedModeKey)
	if err != nil {
		return nil, err
	}
	disablePRFeedMode, err := parseBoolEnv(envGetter, disablePRFeedModeKey)
	if err != nil {
		return nil, err
	}

	return &Config{
		GitHubToken:          envGetter.Getenv(ghTokenKey),
		FreshRSSURL:          envGetter.Getenv(freshRSSURLKey),
		FreshRSSUser:         envGetter.Getenv(freshRSSUserKey),
		FreshRSSToken:        envGetter.Getenv(freshRSSTokenKey),
		DebugMode:            debugMode,
		SingleRunMode:        singleRunMode,
		DisableRepoFeedMode:  disableRepoFeedMode,
		DisableIssueFeedMode: disableIssueFeedMode,
		DisablePRFeedMode:    disablePRFeedMode,
		HTTPTimeout:          httpTimeout,
	}, nil
}
