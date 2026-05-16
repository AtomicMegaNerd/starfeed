package config

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	expectedCsvFields = 4
)

type Config struct {
	GitHosts      []githost.GitHost `validate:"required,min=1"`
	RSSServer     rss.RSSServer     `validate:"required"`
	DebugMode     bool
	SingleRunMode bool
	HTTPTimeout   time.Duration `validate:"required"`
	Client        *http.Client  `validate:"required"`
}

func NewConfig(envGetter EnvGetter, client *http.Client) (*Config, error) {
	validate := validator.New()

	// Parse optional HTTP timeout, default to 10 seconds
	httpTimeout := 10 * time.Second
	if timeoutStr := envGetter.Getenv(httpTimeoutKey); timeoutStr != "" {
		if timeoutSeconds, err := strconv.Atoi(timeoutStr); err == nil && timeoutSeconds > 0 {
			httpTimeout = time.Duration(timeoutSeconds) * time.Second
		}
	}
	// Set the timeout
	client.Timeout = httpTimeout

	debugMode := false
	debugMode, _ = parseBoolEnv(envGetter, debugModeKey)

	singleRunMode := false
	singleRunMode, _ = parseBoolEnv(envGetter, singleRunModeKey)

	gitHosts := make([]githost.GitHost, 0)

	for ix := 0; ; ix++ {
		gitHostCsv := envGetter.Getenv(fmt.Sprintf("%s%d", gitHostKey, ix))
		if gitHostCsv == "" {
			if ix == 0 {
				return nil, errors.New("must define at least 1 git host")
			}
			break
		}

		parts := strings.SplitN(gitHostCsv, ",", expectedCsvFields)
		if len(parts) != expectedCsvFields {
			return nil, fmt.Errorf("expected csv to have %d parts but it had %d", 4, len(parts))
		}

		hostType := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		baseURL := strings.TrimSpace(parts[2])
		token := strings.TrimSpace(parts[3])

		// This will fail validation on construction if any of these are invalid...
		gitHost, err := githost.NewGitHost(hostType, name, baseURL, token, client)
		if err != nil {
			return nil, err
		}

		gitHosts = append(gitHosts, gitHost)
	}

	rssCsv := envGetter.Getenv(rssServerKey)
	parts := strings.SplitN(rssCsv, ",", expectedCsvFields)
	if len(parts) != expectedCsvFields {
		return nil, fmt.Errorf(
			"expected csv to have %d parts but it had %d", expectedCsvFields, len(parts),
		)
	}

	rssType := strings.TrimSpace(parts[0])
	baseUrl := strings.TrimSpace(parts[1])
	user := strings.TrimSpace(parts[2])
	token := strings.TrimSpace(parts[3])

	rssServer, err := rss.NewFreshRSSFeedManager(rssType, baseUrl, user, token, client)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		GitHosts:      gitHosts,
		RSSServer:     rssServer,
		DebugMode:     debugMode,
		SingleRunMode: singleRunMode,
		HTTPTimeout:   httpTimeout,
		Client:        client,
	}

	if err := validate.Struct(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
