package testutils

import "github.com/stretchr/testify/mock"

// MockSSHClient mocks quiltctl/ssh.Client.
type MockSSHClient struct {
	mock.Mock
}

// Run mocks ssh.Client.Run().
func (m *MockSSHClient) Run(allocatePTY bool, command string) error {
	args := m.Called(allocatePTY, command)
	return args.Error(0)
}

// Close mocks ssh.Client.Close().
func (m *MockSSHClient) Close() error {
	args := m.Called()
	return args.Error(0)
}
