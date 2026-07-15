package config

import (
	"os"
)

const (
	configPathEnvVar  = "STARFEED_CONFIG_PATH"
	defaultConfigPath = "./starfeed.toml"
)

type ConfigLoader struct{}

func (cl ConfigLoader) LoadConfig() ([]byte, error) {
	cfgPath := os.Getenv(configPathEnvVar)
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}
	return os.ReadFile(cfgPath)
}
