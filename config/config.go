package config

import (
	"strconv"
	"time"

	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/go-playground/validator/v10"
)

const (
	// required
	gitForgeKey  = "STARFEED_GIT_FORGE"
	rssServerKey = "STARFEED_RSS_SERVER"

	// optional
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey   = "STARFEED_HTTP_TIMEOUT"

	gitForgeConfigFields  = 6
	rssServerConfigFields = 5

	defaultHTTPTimeoutSeconds = 60
)

// The main Config struct used to hold configuration state for the app
type Config struct {
	GitForgeConfigs []gitforge.GitForgeConfig `validate:"required,min=1"`
	RSSServerConfig rss.RSSServerConfig       `validate:"required"`
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

	gitForgeConfigs, err := buildGitForgeConfigs(validate, g)
	if err != nil {
		return Config{}, err
	}
	rssConfig, err := buildRssServerConfig(validate, g)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		GitForgeConfigs: gitForgeConfigs,
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
