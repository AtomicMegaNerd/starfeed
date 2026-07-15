package config

import (
	"bytes"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pelletier/go-toml/v2"
)

const (
	configPathEnvVar  = "STARFEED_CONFIG_PATH"
	defaultConfigPath = "./starfeed.toml"
)

// The main Config struct used to hold configuration state for the app
type Config struct {
	GitForges []GitForgeConfig `validate:"required,min=1" toml:"git_forges"`
	RSSServer RSSServerConfig  `validate:"required"       toml:"rss_server"`
	Debug     bool             `validate:"required"       toml:"debug"`
	SingleRun bool             `validate:"required"       toml:"single_run"`
}

// This type both holds and validates the config for a GitForge
type GitForgeConfig struct {
	Type     string `validate:"required,oneof=github forgejo" toml:"type"`
	Name     string `validate:"required,min=3"                toml:"name"`
	Fqdn     string `validate:"required,min=8"                toml:"fqdn"`
	TokenEnv string `validate:"required"                      toml:"token_env"`

	// This is not loaded from the toml
	Token string `validate:"required,min=10"`
}

// This type both holds and validates the config for the RSS Server
type RSSServerConfig struct {
	Name     string `validate:"required,oneof=freshrss" toml:"name"`
	URL      string `validate:"required,url"            toml:"url"`
	User     string `validate:"required,min=3"          toml:"user"`
	TokenEnv string `validate:"required"                toml:"token_env"`

	// This is not loaded from the toml
	Token string `validate:"required,min=10"`
}

type envGetter interface {
	Getenv(key string) string
}

func NewConfig(g envGetter) (Config, error) {
	validate := validator.New()

	cfgData, err := getConfigurationData(g)
	if err != nil {
		return Config{}, fmt.Errorf(
			"could not load config TOML file %s: %w",
			defaultConfigPath,
			err,
		)
	}

	var cfg Config
	// We are making it strict to disallow unknown fields. This will protect against typos
	dec := toml.NewDecoder(bytes.NewReader(cfgData))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf(
			"could not parse invalid TOML file %s: %w",
			defaultConfigPath,
			err,
		)
	}

	// Load the secrets from the environment
	for ix := range cfg.GitForges {
		cfg.GitForges[ix].Token = g.Getenv(cfg.GitForges[ix].TokenEnv)
	}
	cfg.RSSServer.Token = g.Getenv(cfg.RSSServer.TokenEnv)

	// If anything doesn't load properly (secrets included) this will catch it and fail
	if err := validate.Struct(cfg); err != nil {
		return Config{}, fmt.Errorf("config failed validation: %w", err)
	}

	return cfg, nil
}
