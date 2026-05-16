package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/atomicmeganerd/starfeed/githost"
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
	fRSSURLKey       = "STARFEED_FRESHRSS_URL"
	fRSSUserKey      = "STARFEED_FRESHRSS_USER"
	fRSSTokenKey     = "STARFEED_FRESHRSS_API_TOKEN"
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey   = "STARFEED_HTTP_TIMEOUT"
)

type Config struct {
	GitHosts      []githost.GitHostConfig
	FreshRSSURL   string
	FreshRSSUser  string
	FreshRSSToken string // WARNING: Never log this secret
	DebugMode     bool
	SingleRunMode bool
	HTTPTimeout   time.Duration
}

func NewConfig(envGetter EnvGetter) (*Config, error) {
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

	gitHosts := make([]githost.GitHostConfig, 0)

	for ix := 0; ; ix++ {
		host := envGetter.Getenv(fmt.Sprintf("STARFEED_GIT_HOST_%d", ix))
		if host == "" {
			break
		}

		parts := strings.SplitN(host, ",", 3)

		// Make sure we get 3 parts
		if len(parts) != 3 {
			return nil, fmt.Errorf("STARFEED_GIT_HOST_%d config invalid", ix)
		}

		// Make sure all are valid
		// TODO: Check the URL field to make sure it is valid URL
		// TODO: Check token length
		gitHostType := githost.GitHostType(strings.TrimSpace(parts[0]))
		if !gitHostType.Valid() || parts[1] == "" || parts[2] == "" {
			return nil, fmt.Errorf("STARFEED_GIT_HOST_%d config invalid", ix)
		}

		gitHost := githost.GitHostConfig{
			Type:    gitHostType,
			BaseURL: strings.TrimSpace(parts[1]),
			Token:   strings.TrimSpace(parts[2]),
		}
		gitHosts = append(gitHosts, gitHost)
	}

	cfg := &Config{
		GitHosts:      gitHosts,
		FreshRSSURL:   envGetter.Getenv(fRSSURLKey),
		FreshRSSUser:  envGetter.Getenv(fRSSUserKey),
		FreshRSSToken: envGetter.Getenv(fRSSTokenKey),
		DebugMode:     debugMode,
		SingleRunMode: singleRunMode,
		HTTPTimeout:   httpTimeout,
	}

	if len(cfg.GitHosts) < 1 ||
		cfg.FreshRSSURL == "" ||
		cfg.FreshRSSUser == "" ||
		cfg.FreshRSSToken == "" {
		return nil, errors.New("invalid config, required settings missing")
	}

	return cfg, nil
}
