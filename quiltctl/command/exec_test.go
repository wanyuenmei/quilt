package command

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/NetSys/quilt/api"
	clientMock "github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/quiltctl/ssh"
	"github.com/NetSys/quilt/quiltctl/testutils"
)

type mockExitError struct {
	exitCode int
}

func (err mockExitError) Error() string {
	return fmt.Sprint("exited with error code: ", err.exitCode)
}

func (err mockExitError) ExitStatus() int {
	return err.exitCode
}

func TestExecPTY(t *testing.T) {
	isTerminal = func() bool {
		return true
	}
	workerHost := "worker"
	targetContainer := 1

	mockGetter := new(testutils.Getter)
	mockGetter.On("Client", mock.Anything).Return(&clientMock.Client{}, nil)
	mockGetter.On("ContainerClient", mock.Anything, mock.Anything).Return(
		&clientMock.Client{
			ContainerReturn: []db.Container{
				{
					StitchID: targetContainer,
					DockerID: "foo",
				},
			},
			HostReturn: workerHost,
		}, nil)

	mockSSHClient := new(testutils.MockSSHClient)
	sshGetter := func(host, key string) (ssh.Client, error) {
		assert.Equal(t, workerHost, host)
		assert.Equal(t, "key", key)
		return mockSSHClient, nil
	}
	execCmd := Exec{
		privateKey:      "key",
		command:         "cat /etc/hosts",
		allocatePTY:     true,
		targetContainer: targetContainer,
		sshGetter:       sshGetter,
		clientGetter:    mockGetter,
		common: &commonFlags{
			host: api.DefaultSocket,
		},
	}

	mockSSHClient.On("Run", true, "docker exec -it foo cat /etc/hosts").Return(nil)
	mockSSHClient.On("Close").Return(nil)

	execCmd.Run()

	mockSSHClient.AssertExpectations(t)
}

func TestExecNoPTY(t *testing.T) {
	isTerminal = func() bool {
		return false
	}

	workerHost := "worker"
	targetContainer := 1

	mockGetter := new(testutils.Getter)
	mockGetter.On("Client", mock.Anything).Return(&clientMock.Client{}, nil)
	mockGetter.On("ContainerClient", mock.Anything, mock.Anything).Return(
		&clientMock.Client{
			ContainerReturn: []db.Container{
				{
					StitchID: targetContainer,
					DockerID: "foo",
				},
			},
			HostReturn: workerHost,
		}, nil)

	mockSSHClient := new(testutils.MockSSHClient)
	sshGetter := func(host, key string) (ssh.Client, error) {
		assert.Equal(t, workerHost, host)
		return mockSSHClient, nil
	}
	execCmd := Exec{
		command:         "cat /etc/hosts",
		targetContainer: targetContainer,
		sshGetter:       sshGetter,
		clientGetter:    mockGetter,
		common: &commonFlags{
			host: api.DefaultSocket,
		},
	}

	mockSSHClient.On("Run", false, "docker exec  foo cat /etc/hosts").Return(nil)
	mockSSHClient.On("Close").Return(nil)

	execCmd.Run()

	mockSSHClient.AssertExpectations(t)
}

func TestExecPTYError(t *testing.T) {
	isTerminal = func() bool {
		return false
	}
	exec := Exec{allocatePTY: true}
	exitCode := exec.Run()
	assert.NotEqual(t, 0, exitCode)
}

func TestExecFlags(t *testing.T) {
	t.Parallel()

	checkExecParsing(t, []string{"1", "sh"},
		Exec{
			targetContainer: 1,
			command:         "sh",
		}, nil)
	checkExecParsing(t, []string{"-i", "key", "1", "sh"},
		Exec{
			targetContainer: 1,
			privateKey:      "key",
			command:         "sh",
		}, nil)
	checkExecParsing(t, []string{"1", "cat /etc/hosts"},
		Exec{
			targetContainer: 1,
			command:         "cat /etc/hosts",
		}, nil)
	checkExecParsing(t, []string{"-t", "1", "cat /etc/hosts"},
		Exec{
			allocatePTY:     true,
			targetContainer: 1,
			command:         "cat /etc/hosts",
		}, nil)
	checkExecParsing(t, []string{"1"}, Exec{},
		errors.New("must specify a target container and command"))
	checkExecParsing(t, []string{}, Exec{},
		errors.New("must specify a target container and command"))
}

func checkExecParsing(t *testing.T, args []string, expArgs Exec, expErr error) {

	execCmd := NewExecCommand()
	err := parseHelper(execCmd, args)

	assert.Equal(t, expErr, err)
	assert.Equal(t, expArgs.targetContainer, execCmd.targetContainer)
	assert.Equal(t, expArgs.command, execCmd.command)
	assert.Equal(t, expArgs.privateKey, execCmd.privateKey)
	assert.Equal(t, expArgs.allocatePTY, execCmd.allocatePTY)
}

func TestExecHelperErrors(t *testing.T) {
	// SSH ExitError
	mockSSHClient := new(testutils.MockSSHClient)
	mockSSHClient.On("Run", true, "docker exec -it id123 foo bar").
		Return(mockExitError{10}).Once()

	exitCode := execHelper(mockSSHClient, "id123", "foo bar", true)
	assert.Equal(t, 10, exitCode)
	mockSSHClient.AssertExpectations(t)

	// Other SSH error
	mockSSHClient = new(testutils.MockSSHClient)
	mockSSHClient.On("Run", true, "docker exec -it id123 echo hello").
		Return(assert.AnError).Once()

	exitCode = execHelper(mockSSHClient, "id123", "echo hello", true)
	assert.Equal(t, 1, exitCode)
	mockSSHClient.AssertExpectations(t)
}

func TestExecHelper(t *testing.T) {
	mockSSHClient := new(testutils.MockSSHClient)
	mockSSHClient.On("Run", false, "docker exec  id123 /bin/sh").Return(nil).Once()

	exitCode := execHelper(mockSSHClient, "id123", "/bin/sh", false)
	assert.Equal(t, 0, exitCode)
	mockSSHClient.AssertExpectations(t)
}
