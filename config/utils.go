package config

import (
	"fmt"
	"os"
	"time"
)

const (
	configPathEnvVar  = "STARFEED_CONFIG_PATH"
	defaultConfigPath = "./starfeed.toml"
	minDuration       = 1 * time.Hour
	maxDuration       = 24 * 7 * time.Hour
)

// We can use this internal custom type to enable easy unmarshalling by go-toml
type duration time.Duration

func (d *duration) UnmarshalText(text []byte) error {
	parsed, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	if parsed < minDuration || parsed > maxDuration {
		return fmt.Errorf(
			"field must be set with duration between %s and %s",
			minDuration, maxDuration,
		)
	}
	*d = duration(parsed)
	return nil
}

// This interface lets us mock our ConfigLoader for testing
type configLoader interface {
	LoadConfig() ([]byte, error)
}

// This object is responsible for loading the config file for the application. This concrete
// version loads from disk at the defined path.
type ConfigLoader struct{}

func (cl ConfigLoader) LoadConfig() ([]byte, error) {
	cfgPath := os.Getenv(configPathEnvVar)
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}
	return os.ReadFile(cfgPath)
}
