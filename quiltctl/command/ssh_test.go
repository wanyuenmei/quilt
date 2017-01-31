package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/quiltctl/ssh"
)

func checkSSHParsing(t *testing.T, args []string, expMachine string,
	expSSHArgs []string, expErr error) {

	sshCmd := NewSSHCommand()
	err := parseHelper(sshCmd, args)

	assert.Equal(t, expErr, err)
	assert.Equal(t, expMachine, sshCmd.targetMachine)
	assert.Equal(t, expSSHArgs, sshCmd.sshArgs)
}

func TestSSHFlags(t *testing.T) {
	t.Parallel()

	checkSSHParsing(t, []string{"1"}, "1", []string{}, nil)
	sshArgs := []string{"-i", "~/.ssh/key"}
	checkSSHParsing(t, append([]string{"1"}, sshArgs...), "1", sshArgs, nil)
	checkSSHParsing(t, []string{}, "", nil,
		errors.New("must specify a target machine"))
}

type getMachineTest struct {
	machines   []db.Machine
	query      string
	expErr     string
	expMachine string
}

func TestGetMachine(t *testing.T) {
	t.Parallel()

	tests := []getMachineTest{
		{
			machines: []db.Machine{
				{
					StitchID: "abc",
				},
			},
			query:      "abc",
			expMachine: "abc",
		},
		{
			machines: []db.Machine{
				{
					StitchID: "abc",
				},
				{
					StitchID: "acd",
				},
			},
			query:      "ab",
			expMachine: "abc",
		},
		{
			machines: []db.Machine{
				{
					StitchID: "abc",
				},
				{
					StitchID: "abd",
				},
			},
			query:  "ab",
			expErr: "ambiguous stitchIDs abc and abd",
		},
		{
			machines: []db.Machine{
				{
					StitchID: "abc",
				},
				{
					StitchID: "abd",
				},
			},
			query:  "c",
			expErr: `no machine with stitchID "c"`,
		},
	}
	for _, test := range tests {
		mockClient := mocks.Client{
			MachineReturn: test.machines,
		}
		m, err := getMachine(&mockClient, test.query)

		if test.expErr != "" {
			assert.EqualError(t, err, test.expErr)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, test.expMachine, m.StitchID)
	}
}

type sshTest struct {
	cmd         SSH
	machines    []db.Machine
	expHost     string
	expUseShell bool
	expRunArgs  string
}

func TestSSH(t *testing.T) {
	isTerminal = func() bool { return true }
	tests := []sshTest{
		{
			cmd: SSH{
				common:        &commonFlags{},
				privateKey:    "key",
				targetMachine: "tgt",
			},
			machines:    []db.Machine{{StitchID: "tgt", PublicIP: "host"}},
			expHost:     "host",
			expUseShell: true,
		},
		{
			cmd: SSH{
				common:        &commonFlags{},
				privateKey:    "key",
				targetMachine: "tgt",
				sshArgs:       []string{"foo", "bar"},
			},
			machines:   []db.Machine{{StitchID: "tgt", PublicIP: "host"}},
			expHost:    "host",
			expRunArgs: "foo bar",
		},
	}
	for _, test := range tests {
		testCmd := test.cmd

		mockSSHClient := new(ssh.MockClient)
		testCmd.sshGetter = func(host string, keyPath string) (
			ssh.Client, error) {
			assert.Equal(t, test.expHost, host)
			assert.Equal(t, testCmd.privateKey, keyPath)
			return mockSSHClient, nil
		}
		mockSSHClient.On("Close").Return(nil)
		if test.expUseShell {
			mockSSHClient.On("Shell").Return(nil)
		} else {
			mockSSHClient.On("Run", test.cmd.allocatePTY, test.expRunArgs).
				Return(nil)
		}

		mockQuiltClient := &mocks.Client{
			MachineReturn: test.machines,
		}
		mockClientGetter := new(mocks.Getter)
		mockClientGetter.On("Client", mock.Anything).Return(mockQuiltClient, nil)
		testCmd.clientGetter = mockClientGetter

		assert.Equal(t, 0, testCmd.Run())
		mockSSHClient.AssertExpectations(t)
	}
}
