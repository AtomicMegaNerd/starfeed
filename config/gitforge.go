package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/go-playground/validator/v10"
)

func buildGitForgeConfigs(
	validate *validator.Validate,
	envGetter envGetter,
) ([]gitforge.GitForgeConfig, error) {
	gitForgeConfigs := make([]gitforge.GitForgeConfig, 0)

	for ix := 0; ; ix++ {
		gitForgeCsv := envGetter.Getenv(fmt.Sprintf("%s_%d", gitForgeKey, ix))
		if gitForgeCsv == "" {
			if ix == 0 {
				return nil, errors.New("must define at least 1 git host")
			}
			break
		}

		parts := strings.SplitN(gitForgeCsv, ",", gitForgeConfigFields)
		if len(parts) != gitForgeConfigFields {
			return nil, fmt.Errorf(
				"expected csv to have %d parts but it had %d",
				gitForgeConfigFields,
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

		gitForgeConfig := gitforge.GitForgeConfig{
			Type:    hostType,
			Name:    name,
			BaseURL: baseURL,
			ApiURL:  apiURL,
			Token:   token,
			Enabled: enabled,
		}

		if err := validate.Struct(gitForgeConfig); err != nil {
			return nil, err
		}
		gitForgeConfigs = append(gitForgeConfigs, gitForgeConfig)
	}

	return gitForgeConfigs, nil
}
