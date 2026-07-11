package config

import (
	"fmt"
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

	rssConfig := rss.RSSServerConfig{
		Name:    rssType,
		BaseURL: baseURL,
		User:    user,
		Token:   token,
	}

	if err := validate.Struct(rssConfig); err != nil {
		return rss.RSSServerConfig{}, err
	}

	return rssConfig, nil
}
