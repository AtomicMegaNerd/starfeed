package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type EnvGetter interface {
	Getenv(key string) string
	Getbool(key string) (bool, error)
}

type OSEnvGetter struct{}

func (o OSEnvGetter) Getenv(key string) string {
	return os.Getenv(key)
}

func (o OSEnvGetter) Getbool(key string) (bool, error) {
	if key == "" {
		return false, nil
	}
	return strconv.ParseBool(key)
}

const (
	ghTokenKey       = "STARFEED_GITHUB_API_TOKEN"
	freshRSSURLKey   = "STARFEED_FRESHRSS_URL"
	freshRSSUserKey  = "STARFEED_FRESHRSS_USER"
	freshRSSTokenKey = "STARFEED_FRESHRSS_API_TOKEN"
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey   = "STARFEED_HTTP_TIMEOUT"
)

type Config struct {
	GitHubToken   string
	FreshRSSURL   string
	FreshRSSUser  string
	FreshRSSToken string
	DebugMode     bool
	SingleRunMode bool
	HttpTimeout   time.Duration
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

	debugMode, err := envGetter.Getbool(debugModeKey)
	if err != nil {
		return nil, err
	}
	singleRunMode, err := envGetter.Getbool(singleRunModeKey)
	if err != nil {
		return nil, err
	}

	return &Config{
		GitHubToken:   envGetter.Getenv(ghTokenKey),
		FreshRSSURL:   envGetter.Getenv(freshRSSURLKey),
		FreshRSSUser:  envGetter.Getenv(freshRSSUserKey),
		FreshRSSToken: envGetter.Getenv(freshRSSTokenKey),
		DebugMode:     debugMode,
		SingleRunMode: singleRunMode,
		HttpTimeout:   httpTimeout,
	}, nil
}
