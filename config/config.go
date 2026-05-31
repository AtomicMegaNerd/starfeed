package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	// required
	gitHostKey   = "STARFEED_GIT_HOST_"
	rssServerKey = "STARFEED_RSS_SERVER"

	// optional
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey   = "STARFEED_HTTP_TIMEOUT"

	gitHostConfigFields   = 6
	rssServerConfigFields = 5

	defaultHTTPTimeoutSeconds = 60

	GitHubHostType  = "github"
	ForgejoHostType = "forgejo"
)

// The main Config struct used to hold configuration state for the app
type Config struct {
	GitHostConfigs  []GitHostConfig `validate:"required,min=1"`
	RSSServerConfig RSSServerConfig `validate:"required"`
	DebugMode       bool
	SingleRunMode   bool
	HTTPTimeout     time.Duration `validate:"required"`
}

type envGetter interface {
	Getenv(key string) string
}

func NewConfig(g envGetter) (Config, error) {
	validate := validator.New()

	// Parse optional HTTP timeout
	httpTimeout := defaultHTTPTimeoutSeconds * time.Second
	if timeoutStr := g.Getenv(httpTimeoutKey); timeoutStr != "" {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			httpTimeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	debugMode, err := parseBoolEnv(g, debugModeKey)
	if err != nil {
		return Config{}, err
	}
	singleRunMode, err := parseBoolEnv(g, singleRunModeKey)
	if err != nil {
		return Config{}, err
	}

	gitHostConfigs, err := buildGitHostConfigs(validate, g)
	if err != nil {
		return Config{}, err
	}
	rssConfig, err := buildRssServerConfig(validate, g)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		GitHostConfigs:  gitHostConfigs,
		RSSServerConfig: rssConfig,
		DebugMode:       debugMode,
		SingleRunMode:   singleRunMode,
		HTTPTimeout:     httpTimeout,
	}

	if err := validate.Struct(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// This type both holds and validates the config for a GitHost
type GitHostConfig struct {
	Type    string `validate:"required,oneof=github forgejo"`
	Name    string `validate:"required,min=3"`
	BaseURL string `validate:"required,url"`
	ApiURL  string `validate:"required,url"`
	Token   string `validate:"required,min=10"`
	Enabled bool
}

func buildGitHostConfigs(
	validate *validator.Validate,
	envGetter envGetter,
) ([]GitHostConfig, error) {
	gitHostConfigs := make([]GitHostConfig, 0)

	for ix := 0; ; ix++ {
		gitHostCsv := envGetter.Getenv(fmt.Sprintf("%s%d", gitHostKey, ix))
		if gitHostCsv == "" {
			if ix == 0 {
				return nil, errors.New("must define at least 1 git host")
			}
			break
		}

		parts := strings.SplitN(gitHostCsv, ",", gitHostConfigFields)
		if len(parts) != gitHostConfigFields {
			return nil, fmt.Errorf(
				"expected csv to have %d parts but it had %d",
				gitHostConfigFields,
				len(parts),
			)
		}

		hostType := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		baseURL := strings.TrimSpace(parts[2])
		apiURL := strings.TrimSpace(parts[3])
		token := strings.TrimSpace(parts[4])
		enabledStr := strings.TrimSpace(parts[5])

		enabled, err := strconv.ParseBool(enabledStr)
		if err != nil {
			return nil, fmt.Errorf("invalid Enabled value %q: %w", enabledStr, err)
		}

		gitHostConfig := GitHostConfig{hostType, name, baseURL, apiURL, token, enabled}

		if err := validate.Struct(gitHostConfig); err != nil {
			return nil, err
		}
		gitHostConfigs = append(gitHostConfigs, gitHostConfig)
	}

	return gitHostConfigs, nil
}

// This type both holds and validates the config for the RSS Server
type RSSServerConfig struct {
	Type    string `validate:"required,oneof=freshrss"`
	BaseURL string `validate:"required,url"`
	User    string `validate:"required,min=3"`
	Token   string `validate:"required,min=10"`
	Enabled bool
}

func buildRssServerConfig(
	validate *validator.Validate,
	envGetter envGetter,
) (RSSServerConfig, error) {
	rssCsv := envGetter.Getenv(rssServerKey)

	parts := strings.SplitN(rssCsv, ",", rssServerConfigFields)
	if len(parts) != rssServerConfigFields {
		return RSSServerConfig{}, fmt.Errorf(
			"expected csv to have %d parts but it had %d", rssServerConfigFields, len(parts),
		)
	}
	rssType := strings.TrimSpace(parts[0])
	baseURL := strings.TrimSpace(parts[1])
	user := strings.TrimSpace(parts[2])
	token := strings.TrimSpace(parts[3])
	enabledStr := strings.TrimSpace(parts[4])

	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		return RSSServerConfig{}, fmt.Errorf("invalid Enabled value %q: %w", enabledStr, err)
	}

	rssConfig := RSSServerConfig{rssType, baseURL, user, token, enabled}

	if err := validate.Struct(rssConfig); err != nil {
		return RSSServerConfig{}, err
	}

	return rssConfig, nil
}
