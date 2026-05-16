package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/go-playground/validator/v10"
)

const (
	gitHostKey       = "STARFEED_GIT_HOST_"
	rssServerKey     = "STARFEED_RSS_SERVER"
	debugModeKey     = "STARFEED_DEBUG_MODE"
	singleRunModeKey = "STARFEED_SINGLE_RUN_MODE"
	httpTimeoutKey   = "STARFEED_HTTP_TIMEOUT"
)

type Config struct {
	GitHosts      []githost.GitHostConfig `validate:"required,min=1"`
	RSSServer     rss.RSSServerConfig
	DebugMode     bool
	SingleRunMode bool
	HTTPTimeout   time.Duration
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

	debugMode, err := parseBoolEnv(envGetter, debugModeKey)
	if err != nil {
		return nil, err
	}
	singleRunMode, err := parseBoolEnv(envGetter, singleRunModeKey)
	if err != nil {
		return nil, err
	}

	hostConfigs := make([]githost.GitHostConfig, 0)

	for ix := 0; ; ix++ {
		gitHostCsv := envGetter.Getenv(fmt.Sprintf("%s%d", gitHostKey, ix))
		if gitHostCsv == "" {
			break
		}

		hostConfig, err := githost.ParseGitHostConfigFromCsv(gitHostCsv)
		if err != nil {
			return nil, err
		}

		hostConfigs = append(hostConfigs, *hostConfig)
	}

	rssConfig, err := rss.ParseRSSServerConfigFromCSV(envGetter.Getenv(rssServerKey))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		GitHosts:      hostConfigs,
		RSSServer:     *rssConfig,
		DebugMode:     debugMode,
		SingleRunMode: singleRunMode,
		HTTPTimeout:   httpTimeout,
	}

	return cfg, validate.Struct(cfg)
}
