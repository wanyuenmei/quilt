package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/quilt/quilt/api/client/mocks"
)

func TestVersionFlags(t *testing.T) {
	t.Parallel()

	expHost := "mockHost"

	cmd := NewVersionCommand()
	err := parseHelper(cmd, []string{"-H", expHost})

	assert.NoError(t, err)
	assert.Equal(t, expHost, cmd.common.host)
}

func TestGetDaemonVersion(t *testing.T) {
	t.Parallel()

	mockLocalClient := &mocks.Client{
		VersionReturn: "mockVersion",
	}
	mockGetter := new(mocks.Getter)
	mockGetter.On("Client", mock.Anything).Return(mockLocalClient, nil)

	actual, err := Version{
		clientGetter: mockGetter,
		common:       &commonFlags{},
	}.getDaemonVersion()
	assert.NoError(t, err)
	assert.Equal(t, "mockVersion", actual)

	mockLocalClient.VersionErr = assert.AnError
	_, err = Version{
		clientGetter: mockGetter,
		common:       &commonFlags{},
	}.getDaemonVersion()
	assert.NotNil(t, err)

	mockGetter = new(mocks.Getter)
	mockGetter.On("Client", mock.Anything).Return(nil, assert.AnError)
	_, err = Version{
		clientGetter: mockGetter,
		common:       &commonFlags{},
	}.getDaemonVersion()
	assert.NotNil(t, err)
}
