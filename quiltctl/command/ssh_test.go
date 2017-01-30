package command

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
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

func TestSSHCommandCreation(t *testing.T) {
	t.Parallel()

	exp := []string{"ssh", "quilt@host", "-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", "-i", "~/.ssh/quilt"}
	res := runSSHCommand("host", []string{"-i", "~/.ssh/quilt"})

	assert.Equal(t, exp, res.Args)
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
