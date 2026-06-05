package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	FreshRSSType = "freshrss"
)

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
