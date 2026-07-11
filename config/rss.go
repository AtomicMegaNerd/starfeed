package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/go-playground/validator/v10"
)

func buildRssServerConfig(
	validate *validator.Validate,
	envGetter envGetter,
) (rss.RSSServerConfig, error) {
	rssCsv := envGetter.Getenv(rssServerKey)

	parts := strings.SplitN(rssCsv, ",", rssServerConfigFields)
	if len(parts) != rssServerConfigFields {
		return rss.RSSServerConfig{}, fmt.Errorf(
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
		return rss.RSSServerConfig{}, fmt.Errorf("invalid Enabled value %q: %w", enabledStr, err)
	}

	rssConfig := rss.RSSServerConfig{
		Type:    rssType,
		BaseURL: baseURL,
		User:    user,
		Token:   token,
		Enabled: enabled,
	}

	if err := validate.Struct(rssConfig); err != nil {
		return rss.RSSServerConfig{}, err
	}

	return rssConfig, nil
}
