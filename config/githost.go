package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	GitHubHostType  = "github"
	ForgejoHostType = "forgejo"
)

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
