package command

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quilt/quilt/api/client/mocks"
)

func TestSetupClient(t *testing.T) {
	t.Parallel()

	mockGetter := &mocks.Getter{}

	// Test that we obtain a client, and properly save it.
	cmd := connectionHelper{
		connectionFlags: connectionFlags{
			host: "host",
		},
	}
	expClient := &mocks.Client{}
	mockGetter.On("Client", "host").Return(expClient, nil).Once()
	err := cmd.setupClient(mockGetter)
	assert.NoError(t, err)
	assert.Equal(t, expClient, cmd.client)

	// Test that errors obtaining a client are properly propagated.
	cmd = connectionHelper{
		connectionFlags: connectionFlags{
			host: "host",
		},
	}
	mockGetter.On("Client", "host").Return(nil, assert.AnError).Once()
	err = cmd.setupClient(mockGetter)
	assert.NotNil(t, err)
}
