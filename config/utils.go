package config

import (
	"os"
)

type OSEnvGetter struct{}

func (o OSEnvGetter) Getenv(key string) string {
	return os.Getenv(key)
}

func getConfigurationData(g envGetter) ([]byte, error) {
	cfgPath := g.Getenv(configPathEnvVar)
	if cfgPath == "" {
		cfgPath = defaultConfigPath
	}

	return os.ReadFile(cfgPath)
}
