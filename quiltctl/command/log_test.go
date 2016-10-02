package command

import (
	"errors"
	"os/exec"
	"reflect"
	"testing"

	"github.com/NetSys/quilt/api"
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
	targetContainer := 1
	logsCmd := Log{
		privateKey:      "key",
		targetContainer: targetContainer,
		host:            api.DefaultSocket,
	}

	expArgs := []string{"-t", "-i", "key", "docker logs foo"}
	var sshCalled bool
	ssh = func(host string, args []string) *exec.Cmd {
		sshCalled = true
		if host != "worker" {
			t.Errorf("Bad ssh host: expected %s, but got %s",
				"worker", host)
		}
		if !reflect.DeepEqual(args, expArgs) {
			t.Errorf("Bad ssh args: expected %v, but got %v", expArgs, args)
		}
		return &exec.Cmd{}
	}

	logsCmd.Run()

	if !sshCalled {
		t.Errorf("Never tried SSHing")
	}
}

func TestLogOptions(t *testing.T) {
	targetContainer := 1
	logsCmd := Log{
		privateKey:      "key",
		targetContainer: targetContainer,
		host:            api.DefaultSocket,
		shouldTail:      true,
		showTimestamps:  true,
		sinceTimestamp:  "2006-01-02T15:04:05",
	}

	expArgs := []string{"-t", "-i", "key", "docker logs " +
		"--since=2006-01-02T15:04:05 --timestamps --follow foo"}
	var sshCalled bool
	ssh = func(host string, args []string) *exec.Cmd {
		sshCalled = true
		if host != "worker" {
			t.Errorf("Bad ssh host: expected %s, but got %s",
				"worker", host)
		}
		if !reflect.DeepEqual(args, expArgs) {
			t.Errorf("Bad ssh args: expected %v, but got %v", expArgs, args)
		}
		return &exec.Cmd{}
	}

	logsCmd.Run()

	if !sshCalled {
		t.Errorf("Never tried SSHing")
	}
}

func checkLogParsing(t *testing.T, args []string, exp Log, expErr error) {
	logsCmd := Log{}
	err := logsCmd.Parse(args)

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
