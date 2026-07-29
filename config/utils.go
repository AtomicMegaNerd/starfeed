package config

import (
	"os"
)

const (
	configPathEnvVar  = "STARFEED_CONFIG_PATH"
	defaultConfigPath = "./starfeed.toml"
)

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
