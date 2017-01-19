package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/NetSys/quilt/api"
	clientMock "github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/quiltctl/ssh"
	"github.com/NetSys/quilt/quiltctl/testutils"
)

func TestLogFlags(t *testing.T) {
	t.Parallel()

	checkLogParsing(t, []string{"1"}, Log{
		targetContainer: "1",
	}, nil)
	checkLogParsing(t, []string{"-i", "key", "1"}, Log{
		targetContainer: "1",
		privateKey:      "key",
	}, nil)
	checkLogParsing(t, []string{"-f", "1"}, Log{
		targetContainer: "1",
		shouldTail:      true,
	}, nil)
	checkLogParsing(t, []string{"-t", "1"}, Log{
		targetContainer: "1",
		showTimestamps:  true,
	}, nil)
	checkLogParsing(t, []string{"--since=07/27/2016", "1"}, Log{
		targetContainer: "1",
		sinceTimestamp:  "07/27/2016",
	}, nil)
	checkLogParsing(t, []string{}, Log{},
		errors.New("must specify a target container"))
}

func TestLog(t *testing.T) {
	t.Parallel()

	workerHost := "worker"
	targetContainer := "1"

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
	logsCmd := Log{
		privateKey:      "key",
		targetContainer: targetContainer,
		shouldTail:      true,
		showTimestamps:  true,
		sinceTimestamp:  "2006-01-02T15:04:05",
		sshGetter:       sshGetter,
		clientGetter:    mockGetter,
		common: &commonFlags{
			host: api.DefaultSocket,
		},
	}

	mockSSHClient.On("Run", false, "docker logs --since=2006-01-02T15:04:05 "+
		"--timestamps --follow foo").Return(nil)
	mockSSHClient.On("Close").Return(nil)

	logsCmd.Run()

	mockSSHClient.AssertExpectations(t)
}

func checkLogParsing(t *testing.T, args []string, exp Log, expErr error) {
	logsCmd := NewLogCommand()
	err := parseHelper(logsCmd, args)

	assert.Equal(t, expErr, err)
	assert.Equal(t, exp.targetContainer, logsCmd.targetContainer)
	assert.Equal(t, exp.privateKey, logsCmd.privateKey)
	assert.Equal(t, exp.sinceTimestamp, logsCmd.sinceTimestamp)
	assert.Equal(t, exp.showTimestamps, logsCmd.showTimestamps)
	assert.Equal(t, exp.shouldTail, logsCmd.shouldTail)
}
