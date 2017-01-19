package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/quiltctl/ssh"
	"github.com/NetSys/quilt/quiltctl/testutils"
)

func TestLoginParsing(t *testing.T) {
	t.Parallel()

	checkLoginParsing(t, []string{}, Login{}, "must specify a target container")
	checkLoginParsing(t, []string{"-i", "key"}, Login{},
		"must specify a target container")

	expArgs := Login{
		targetContainer: "12",
	}
	checkLoginParsing(t, []string{"12"}, expArgs, "")

	expArgs = Login{
		privateKey:      "key",
		targetContainer: "13",
	}
	checkLoginParsing(t, []string{"-i", "key", "13"}, expArgs, "")
}

func checkLoginParsing(t *testing.T, args []string, expArgs Login, expErrMsg string) {
	loginCmd := NewLoginCommand()
	err := parseHelper(loginCmd, args)

	if expErrMsg != "" {
		assert.EqualError(t, err, expErrMsg)
		return
	}

	assert.Nil(t, err, "did not expect an error")
	assert.Equal(t, expArgs.targetContainer, loginCmd.targetContainer)
	assert.Equal(t, expArgs.privateKey, loginCmd.privateKey)
}

func TestLoginRunErrors(t *testing.T) {
	var cmd *Login
	var mockClientGetter *testutils.Getter
	var mockSSHGetter ssh.Getter
	var mockLocalClient, mockContainerClient *mocks.Client
	var containers []db.Container

	origIsTerminal := isTerminal
	isTerminal = func() bool { return true }
	defer func() { isTerminal = origIsTerminal }()

	// Fail to get localClient
	mockClientGetter = new(testutils.Getter)
	mockClientGetter.On("Client", mock.Anything).Return(nil, assert.AnError)

	cmd = &Login{common: &commonFlags{}, clientGetter: mockClientGetter}
	assert.Equal(t, 1, cmd.Run())
	mockClientGetter.AssertExpectations(t)

	// Fail to get containerClient
	mockLocalClient = new(mocks.Client)
	mockClientGetter = new(testutils.Getter)
	mockClientGetter.On("Client", mock.Anything).Return(mockLocalClient, nil)
	mockClientGetter.On("ContainerClient", mockLocalClient, mock.Anything).
		Return(nil, assert.AnError)

	cmd = &Login{
		common:       &commonFlags{},
		clientGetter: mockClientGetter,
	}
	assert.Equal(t, 1, cmd.Run())
	mockClientGetter.AssertExpectations(t)

	// Fail to get container information
	mockLocalClient = &mocks.Client{}
	mockContainerClient = &mocks.Client{ContainerErr: assert.AnError}

	mockClientGetter = new(testutils.Getter)
	mockClientGetter.On("Client", mock.Anything).Return(mockLocalClient, nil)
	mockClientGetter.On("ContainerClient", mockLocalClient, mock.Anything).
		Return(mockContainerClient, nil)

	cmd = &Login{common: &commonFlags{}, clientGetter: mockClientGetter}
	assert.Equal(t, 1, cmd.Run())
	mockClientGetter.AssertExpectations(t)

	// Fail to open ssh connection
	mockLocalClient = &mocks.Client{}

	containers = []db.Container{{StitchID: "777"}}
	mockContainerClient = &mocks.Client{ContainerReturn: containers}
	mockSSHGetter = func(host string, keyPath string) (ssh.Client, error) {
		return nil, assert.AnError
	}

	mockClientGetter = new(testutils.Getter)
	mockClientGetter.On("Client", mock.Anything).Return(mockLocalClient, nil)
	mockClientGetter.On("ContainerClient", mockLocalClient, mock.Anything).
		Return(mockContainerClient, nil)

	cmd = &Login{
		common:          &commonFlags{},
		clientGetter:    mockClientGetter,
		sshGetter:       mockSSHGetter,
		targetContainer: "777",
	}
	assert.Equal(t, 1, cmd.Run())
	mockClientGetter.AssertExpectations(t)
}

func TestLoginRunNotIsTerminal(t *testing.T) {
	origIsTerminal := isTerminal
	isTerminal = func() bool { return false }
	defer func() { isTerminal = origIsTerminal }()

	loginCmd := &Login{}

	assert.Equal(t, 1, loginCmd.Run())
}

func TestLoginRunSuccess(t *testing.T) {
	origIsTerminal := isTerminal
	isTerminal = func() bool { return true }
	defer func() { isTerminal = origIsTerminal }()

	mockLocalClient := &mocks.Client{}

	containers := []db.Container{{StitchID: "100"}}
	mockContainerClient := &mocks.Client{ContainerReturn: containers}

	sshClient := new(testutils.MockSSHClient)
	sshClient.On("Run", mock.Anything, mock.Anything).Return(nil)
	sshClient.On("Close").Return(nil).Once()
	mockSSHGetter := func(host string, keyPath string) (ssh.Client, error) {
		return sshClient, nil
	}

	mockClientGetter := new(testutils.Getter)
	mockClientGetter.On("Client", mock.Anything).Return(mockLocalClient, nil)
	mockClientGetter.On("ContainerClient", mockLocalClient, "100").
		Return(mockContainerClient, nil)

	cmd := &Login{
		common:          &commonFlags{},
		clientGetter:    mockClientGetter,
		sshGetter:       mockSSHGetter,
		targetContainer: "100",
	}
	assert.Equal(t, 0, cmd.Run())
	mockClientGetter.AssertExpectations(t)
	sshClient.AssertExpectations(t)
}
