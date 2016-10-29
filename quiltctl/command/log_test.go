package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/NetSys/quilt/api"
	clientMock "github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/quiltctl/testutils"
)

func TestLogFlags(t *testing.T) {
	t.Parallel()

	checkLogParsing(t, []string{"1"}, Log{
		targetContainer: 1,
	}, nil)
	checkLogParsing(t, []string{"-i", "key", "1"}, Log{
		targetContainer: 1,
		privateKey:      "key",
	}, nil)
	checkLogParsing(t, []string{"-f", "1"}, Log{
		targetContainer: 1,
		shouldTail:      true,
	}, nil)
	checkLogParsing(t, []string{"-t", "1"}, Log{
		targetContainer: 1,
		showTimestamps:  true,
	}, nil)
	checkLogParsing(t, []string{"--since=07/27/2016", "1"}, Log{
		targetContainer: 1,
		sinceTimestamp:  "07/27/2016",
	}, nil)
	checkLogParsing(t, []string{}, Log{
		targetContainer: 0,
	}, errors.New("must specify a target container"))
}

func TestLog(t *testing.T) {
	t.Parallel()

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
	logsCmd := Log{
		privateKey:      "key",
		targetContainer: targetContainer,
		shouldTail:      true,
		showTimestamps:  true,
		sinceTimestamp:  "2006-01-02T15:04:05",
		SSHClient:       mockSSHClient,
		clientGetter:    mockGetter,
		common: &commonFlags{
			host: api.DefaultSocket,
		},
	}

	mockSSHClient.On("Connect", workerHost, "key").Return(nil)
	mockSSHClient.On("Run", "docker logs --since=2006-01-02T15:04:05 --timestamps "+
		"--follow foo").Return(nil)
	mockSSHClient.On("Disconnect").Return(nil)

	logsCmd.Run()

	mockSSHClient.AssertExpectations(t)
}

func checkLogParsing(t *testing.T, args []string, exp Log, expErr error) {
	logsCmd := NewLogCommand(nil)
	err := parseHelper(logsCmd, args)

	if err != nil {
		if expErr != nil {
			if err.Error() != expErr.Error() {
				t.Errorf("Expected error %s, but got %s",
					expErr.Error(), err.Error())
			}
			return
		}

		t.Errorf("Unexpected error when parsing log args: %s", err.Error())
		return
	}

	if logsCmd.targetContainer != exp.targetContainer {
		t.Errorf("Expected log command to parse target container %d, but got %d",
			exp.targetContainer, logsCmd.targetContainer)
	}

	if logsCmd.privateKey != exp.privateKey {
		t.Errorf("Expected log command to parse private key %s, but got %s",
			exp.privateKey, logsCmd.privateKey)
	}

	if logsCmd.sinceTimestamp != exp.sinceTimestamp {
		t.Errorf("Expected log command to parse since timestamp %s, but got %s",
			exp.sinceTimestamp, logsCmd.sinceTimestamp)
	}

	if logsCmd.showTimestamps != exp.showTimestamps {
		t.Errorf("Expected log command to parse timestamp flag %t, but got %t",
			exp.showTimestamps, logsCmd.showTimestamps)
	}

	if logsCmd.shouldTail != exp.shouldTail {
		t.Errorf("Expected log command to parse tail flag %t, but got %t",
			exp.shouldTail, logsCmd.shouldTail)
	}
}
