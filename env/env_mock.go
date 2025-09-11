package env

var _ Env = (*mockEnv)(nil)

type mockEnv struct {
	getFunc func(key string) (string, error)
}

func (m *mockEnv) Get(key string) (string, error) {
	return m.getFunc(key)
}
