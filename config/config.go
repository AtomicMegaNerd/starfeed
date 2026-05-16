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
	gitHostKey       = "STARFEED_GIT_HOST_"
	rssServerKey     = "STARFEED_RSS_SERVER"
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey   = "STARFEED_HTTP_TIMEOUT"

	gitHubHostConfigFields = 5
	rssServerConfigFields  = 4
)

type GitHostConfig struct {
	Type    string `validate:"required,oneof=github forgejo"`
	Name    string `validate:"required,min=3"`
	BaseURL string `validate:"required,url"`
	ApiURL  string `validate:"required,url"`
	Token   string `validate:"required,min=24"`
}

type RSSServerConfig struct {
	Type    string `validate:"required,oneof=freshrss"`
	BaseURL string `validate:"required,url"`
	User    string `validate:"required,min=3"`
	Token   string `validate:"required,min=10"`
}

type Config struct {
	GitHostConfigs  []GitHostConfig  `validate:"required,min=1"`
	RSSServerConfig *RSSServerConfig `validate:"required"`
	DebugMode       bool
	SingleRunMode   bool
	HTTPTimeout     time.Duration `validate:"required"`
}

func NewConfig(envGetter EnvGetter) (*Config, error) {
	validate := validator.New()

	// Parse optional HTTP timeout, default to 10 seconds
	httpTimeout := 10 * time.Second
	if timeoutStr := envGetter.Getenv(httpTimeoutKey); timeoutStr != "" {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			httpTimeout = time.Duration(timeoutSeconds) * time.Second
		}
	}

	debugMode := false
	debugMode, _ = parseBoolEnv(envGetter, debugModeKey)
	singleRunMode := false
	singleRunMode, _ = parseBoolEnv(envGetter, singleRunModeKey)

	gitHostConfigs, err := buildGitHostConfigs(validate, envGetter)
	if err != nil {
		return nil, err
	}
	rssConfig, err := buildRssServerConfig(validate, envGetter)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		GitHostConfigs:  gitHostConfigs,
		RSSServerConfig: rssConfig,
		DebugMode:       debugMode,
		SingleRunMode:   singleRunMode,
		HTTPTimeout:     httpTimeout,
	}

	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func buildGitHostConfigs(
	validate *validator.Validate,
	envGetter EnvGetter,
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

		parts := strings.SplitN(gitHostCsv, ",", gitHubHostConfigFields)
		if len(parts) != gitHubHostConfigFields {
			return nil, fmt.Errorf("expected csv to have %d parts but it had %d", 4, len(parts))
		}

		hostType := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		baseURL := strings.TrimSpace(parts[2])
		apiURL := strings.TrimSpace(parts[3])
		token := strings.TrimSpace(parts[4])

		// This will fail validation on construction if any of these are invalid...
		gitHostConfig := GitHostConfig{hostType, name, baseURL, apiURL, token}

		if err := validate.Struct(gitHostConfig); err != nil {
			return nil, err
		}
		gitHostConfigs = append(gitHostConfigs, gitHostConfig)
	}

	return gitHostConfigs, nil
}

func buildRssServerConfig(
	validate *validator.Validate,
	envGetter EnvGetter,
) (*RSSServerConfig, error) {
	rssCsv := envGetter.Getenv(rssServerKey)

	parts := strings.SplitN(rssCsv, ",", rssServerConfigFields)
	if len(parts) != rssServerConfigFields {
		return nil, fmt.Errorf(
			"expected csv to have %d parts but it had %d", rssServerConfigFields, len(parts),
		)
	}
	rssType := strings.TrimSpace(parts[0])
	baseUrl := strings.TrimSpace(parts[1])
	user := strings.TrimSpace(parts[2])
	token := strings.TrimSpace(parts[3])

	rssConfig := &RSSServerConfig{rssType, baseUrl, user, token}

	if err := validate.Struct(rssConfig); err != nil {
		return nil, err
	}

	return rssConfig, nil
}
