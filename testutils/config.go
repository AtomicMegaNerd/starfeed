package testutils

type MockConfigLoader struct {
	ExpectedData  []byte
	ExpectedError error
}

func (cl MockConfigLoader) LoadConfig() ([]byte, error) {
	return cl.ExpectedData, cl.ExpectedError
}
