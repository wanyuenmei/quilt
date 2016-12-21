package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	clientMock "github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/quiltctl/testutils"
)

func TestPsFlags(t *testing.T) {
	t.Parallel()

	expHost := "IP"

	cmd := NewPsCommand()
	err := parseHelper(cmd, []string{"-H", expHost})

	assert.NoError(t, err)
	assert.Equal(t, expHost, cmd.common.host)
}

func TestPsErrors(t *testing.T) {
	t.Parallel()

	var cmd *Ps
	var mockGetter *testutils.Getter
	var mockClient, mockLeaderClient *clientMock.Client

	// Error connecting to local client
	mockGetter = new(testutils.Getter)
	mockGetter.On("Client", mock.Anything).Return(nil, assert.AnError)

	cmd = &Ps{&commonFlags{}, mockGetter}
	assert.Equal(t, 1, cmd.Run())
	mockGetter.AssertExpectations(t)

	// Error querying machines
	mockGetter = new(testutils.Getter)
	mockClient = &clientMock.Client{MachineErr: assert.AnError}
	mockGetter.On("Client", mock.Anything).Return(mockClient, nil)

	cmd = &Ps{&commonFlags{}, mockGetter}
	assert.Equal(t, 1, cmd.Run())
	mockGetter.AssertExpectations(t)

	// Error connecting to leader
	mockGetter = new(testutils.Getter)
	mockClient = new(clientMock.Client)
	mockGetter.On("Client", mock.Anything).Return(mockClient, nil)
	mockGetter.On("LeaderClient", mock.Anything).Return(nil, assert.AnError)

	cmd = &Ps{&commonFlags{}, mockGetter}
	assert.Equal(t, 1, cmd.Run())
	mockGetter.AssertExpectations(t)

	// Error querying containers
	mockGetter = new(testutils.Getter)
	mockClient = new(clientMock.Client)
	mockLeaderClient = &clientMock.Client{ContainerErr: assert.AnError}
	mockGetter.On("Client", mock.Anything).Return(mockClient, nil)
	mockGetter.On("LeaderClient", mock.Anything).Return(mockLeaderClient, nil)

	cmd = &Ps{&commonFlags{}, mockGetter}
	assert.Equal(t, 1, cmd.Run())
	mockGetter.AssertExpectations(t)
}

func TestPsSuccess(t *testing.T) {
	t.Parallel()

	mockGetter := new(testutils.Getter)
	mockClient := new(clientMock.Client)
	mockLeaderClient := new(clientMock.Client)

	mockGetter.On("Client", mock.Anything).Return(mockClient, nil)
	mockGetter.On("LeaderClient", mock.Anything).Return(mockLeaderClient, nil)

	cmd := &Ps{&commonFlags{}, mockGetter}
	assert.Equal(t, 0, cmd.Run())
	mockGetter.AssertExpectations(t)
}
