package config

import (
	"os"
	"strconv"
)

type EnvGetter interface {
	Getenv(key string) string
}

type OSEnvGetter struct{}

func (o OSEnvGetter) Getenv(key string) string {
	return os.Getenv(key)
}

func parseBoolEnv(envGetter EnvGetter, key string) (bool, error) {
	v := envGetter.Getenv(key)
	if v == "" {
		return false, nil
	}
	return strconv.ParseBool(v)
}
