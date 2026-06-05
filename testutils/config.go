package testutils

// MockEnvGetter implements EnvGetter for testing
type MockEnvGetter struct {
	envVars map[string]string
}

func NewMockEnvGetter(envVars map[string]string) *MockEnvGetter {
	return &MockEnvGetter{envVars: envVars}
}

func (m *MockEnvGetter) Getenv(key string) string {
	return m.envVars[key]
}
