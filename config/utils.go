package config

import (
	"os"
	"strconv"
)

type OSEnvGetter struct{}

func (o OSEnvGetter) Getenv(key string) string {
	return os.Getenv(key)
}

func parseBoolEnv(envGetter envGetter, key string) (bool, error) {
	v := envGetter.Getenv(key)
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}
