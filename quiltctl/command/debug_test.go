package command

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/quilt/quilt/api/client/mocks"
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/quiltctl/ssh"
	"github.com/quilt/quilt/util"
)

var debugFolder = "debug_logs_Mon_Jan_01_00-00-00"

func checkDebugParsing(t *testing.T, args []string, expArgs Debug, expErrMsg string) {
	debugCmd := NewDebugCommand()
	err := parseHelper(debugCmd, args)

	if expErrMsg != "" {
		assert.EqualError(t, err, expErrMsg)
		return
	}

	assert.NoError(t, err)
	assert.Equal(t, expArgs.all, debugCmd.all)
	assert.Equal(t, expArgs.containers, debugCmd.containers)
	assert.Equal(t, expArgs.machines, debugCmd.machines)
	assert.Equal(t, expArgs.privateKey, debugCmd.privateKey)
	assert.Equal(t, expArgs.ids, debugCmd.ids)
}

func TestDebugFlags(t *testing.T) {
	t.Parallel()

	checkDebugParsing(t, []string{"-i", "key", "1"},
		Debug{
			privateKey: "key",
			tar:        true,
			ids:        []string{"1"},
		}, "")
	checkDebugParsing(t, []string{"-i", "key", "-all"},
		Debug{
			privateKey: "key",
			tar:        true,
			all:        true,
			ids:        []string{},
		}, "")
	checkDebugParsing(t, []string{"-i", "key", "-containers"},
		Debug{
			privateKey: "key",
			tar:        true,
			containers: true,
			ids:        []string{},
		}, "")
	checkDebugParsing(t, []string{"-i", "key", "-machines"},
		Debug{
			privateKey: "key",
			tar:        true,
			machines:   true,
			ids:        []string{},
		}, "")
	checkDebugParsing(t, []string{"-i", "key", "id1", "id2"},
		Debug{
			privateKey: "key",
			tar:        true,
			ids:        []string{"id1", "id2"},
		}, "")
	checkDebugParsing(t, []string{"-all", "-machines", "id1", "id2"},
		Debug{
			tar:      true,
			all:      true,
			machines: true,
			ids:      []string{"id1", "id2"},
		}, "")
	checkDebugParsing(t, []string{"-containers", "-machines", "id1", "id2"},
		Debug{
			tar:        true,
			containers: true,
			machines:   true,
			ids:        []string{"id1", "id2"},
		}, "")
	checkDebugParsing(t, []string{"-tar=false", "-machines"},
		Debug{
			tar:      false,
			machines: true,
			ids:      []string{},
		}, "")
	checkDebugParsing(t, []string{"-o=tmp_folder", "-machines"},
		Debug{
			tar:      true,
			outPath:  "tmp_folder",
			machines: true,
			ids:      []string{},
		}, "")
	checkDebugParsing(t, []string{}, Debug{},
		"must supply at least one ID or set option")
	checkDebugParsing(t, []string{"-i", "key"}, Debug{},
		"must supply at least one ID or set option")
}

type debugTest struct {
	cmd        Debug
	machines   []db.Machine
	containers []db.Container

	expSSH      bool
	expReturn   int
	expFiles    []string
	expNotFiles []string
}

func TestDebug(t *testing.T) {
	timestamp = func() time.Time {
		return time.Time{}
	}
	defer func() {
		timestamp = time.Now
	}()

	execCmd = func(name string, arg ...string) *exec.Cmd {
		assert.Equal(t, name, "quilt")
		return exec.Command("echo", "hello world")
	}
	defer func() {
		execCmd = exec.Command
	}()

	tests := []debugTest{
		// Check that all logs are fetched.
		{
			cmd: Debug{
				tar:    false,
				all:    true,
				common: &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("1", false, ""),
				append(filesFor("2", true, ""),
					filesFor("3", true, "")...)...),
			expNotFiles: []string{},
		},
		// Check that all logs are fetched with -machines and -containers.
		{
			cmd: Debug{
				tar:        false,
				machines:   true,
				containers: true,
				common:     &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("1", false, ""),
				append(filesFor("2", true, ""),
					filesFor("3", true, "")...)...),
			expNotFiles: []string{},
		},
		// Check that just container logs are fetched.
		{
			cmd: Debug{
				tar:        false,
				containers: true,
				common:     &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("2", true, ""),
				filesFor("3", true, "")...),
			expNotFiles: []string{},
		},
		// Check that just machine logs are fetched.
		{
			cmd: Debug{
				tar:      false,
				machines: true,
				common:   &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
				{
					StitchID:  "4",
					PublicIP:  "5.6.7.8",
					PrivateIP: "8.7.6.5",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "8.7.6.5"},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("1", false, ""),
				filesFor("4", false, "")...),
			expNotFiles: []string{},
		},
		// Check that we can get logs by specific stitch ids
		{
			cmd: Debug{
				tar:    false,
				ids:    []string{"2", "4", "5"},
				common: &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
				{StitchID: "4", DockerID: "c", Minion: "4.3.2.1"},
				{StitchID: "5", DockerID: "d", Minion: "4.3.2.1"},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("2", true, ""),
				append(filesFor("4", true, ""),
					filesFor("5", true, "")...)...),
			expNotFiles: []string{},
		},
		// Check that we can get logs by specific stitch ids in arbitrary order
		{
			cmd: Debug{
				tar:    false,
				ids:    []string{"4", "2", "1"},
				common: &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
				{StitchID: "4", DockerID: "c", Minion: "4.3.2.1"},
				{StitchID: "5", DockerID: "d", Minion: "4.3.2.1"},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("1", false, ""),
				append(filesFor("4", true, ""),
					filesFor("2", true, "")...)...),
			expNotFiles: []string{},
		},
		// Check that we error on arbitrary stitch IDs.
		{
			cmd: Debug{
				tar:    false,
				ids:    []string{"4", "2"},
				common: &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "409",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
				{StitchID: "41", DockerID: "c", Minion: "4.3.2.1"},
				{StitchID: "5", DockerID: "d", Minion: "4.3.2.1"},
			},
			expSSH:      false,
			expReturn:   1,
			expFiles:    []string{},
			expNotFiles: []string{},
		},
		// Check that we error on non-existent stitch IDs.
		{
			cmd: Debug{
				tar:    false,
				ids:    []string{"6"},
				common: &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "409",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
				{StitchID: "41", DockerID: "c", Minion: "4.3.2.1"},
				{StitchID: "5", DockerID: "d", Minion: "4.3.2.1"},
			},
			expSSH:      false,
			expReturn:   1,
			expFiles:    []string{},
			expNotFiles: []string{},
		},
		// Check that containers without a minion aren't reported.
		{
			cmd: Debug{
				tar:        false,
				containers: true,
				common:     &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
				{StitchID: "4", DockerID: "c", Minion: "4.3.2.1"},
				{StitchID: "5", DockerID: "d", Minion: ""},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("2", true, ""),
				append(filesFor("3", true, ""),
					filesFor("4", true, "")...)...),
			expNotFiles: filesFor("5", true, ""),
		},
		// Check that machines without an IP aren't reported.
		{
			cmd: Debug{
				tar:      false,
				machines: true,
				common:   &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
				{
					StitchID:  "4",
					PublicIP:  "",
					PrivateIP: "",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
			},
			expSSH:      true,
			expReturn:   0,
			expFiles:    filesFor("1", false, ""),
			expNotFiles: filesFor("4", false, ""),
		},
		// Check that a supplied path is respected.
		{
			cmd: Debug{
				tar:     false,
				all:     true,
				outPath: "tmp_folder",
				common:  &commonFlags{},
			},
			machines: []db.Machine{
				{
					StitchID:  "1",
					PublicIP:  "1.2.3.4",
					PrivateIP: "4.3.2.1",
					Role:      db.Worker,
				},
			},
			containers: []db.Container{
				{StitchID: "2", DockerID: "a", Minion: "4.3.2.1"},
				{StitchID: "3", DockerID: "b", Minion: "4.3.2.1"},
			},
			expSSH:    true,
			expReturn: 0,
			expFiles: append(filesFor("1", false, "tmp_folder"),
				append(filesFor("2", true, "tmp_folder"),
					filesFor("3", true, "tmp_folder")...)...),
			expNotFiles: append(filesFor("1", false, ""),
				append(filesFor("2", true, ""),
					filesFor("3", true, "")...)...),
		},
	}

	for _, test := range tests {
		util.AppFs = afero.NewMemMapFs()
		testCmd := test.cmd

		mockSSHClient := new(ssh.MockClient)
		testCmd.sshGetter = func(host string, keyPath string) (
			ssh.Client, error) {

			assert.Equal(t, testCmd.privateKey, keyPath)
			return mockSSHClient, nil
		}
		if test.expSSH {
			mockSSHClient.On("CombinedOutput",
				mock.Anything).Return([]byte(""), nil)
		}

		mockLocalClient := &mocks.Client{
			MachineReturn:   test.machines,
			ContainerReturn: test.containers,
		}

		mockClientGetter := new(mocks.Getter)
		mockClientGetter.On("Client", mock.Anything).Return(mockLocalClient, nil)
		testCmd.clientGetter = mockClientGetter

		assert.Equal(t, test.expReturn, testCmd.Run())
		rootDir := debugFolder
		if test.cmd.outPath != "" {
			rootDir = test.cmd.outPath
		}

		// There should only be daemon files if the fetch succeeded and we didn't
		// tarball the results.
		if test.expReturn == 0 && !test.cmd.tar {
			for _, cmd := range daemonCmds {
				file := filepath.Join(rootDir, cmd.name)
				exists, err := util.FileExists(file)
				assert.NoError(t, err)
				assert.True(t, exists)
			}
		}

		for _, f := range test.expFiles {
			exists, err := util.FileExists(f)
			assert.NoError(t, err)
			assert.True(t, exists)
		}

		for _, f := range test.expNotFiles {
			exists, _ := util.FileExists(f)
			assert.False(t, exists)
		}

		mockSSHClient.AssertExpectations(t)
	}
}

func filesFor(id string, container bool, outpath string) []string {
	prefix := machineDir
	cmds := machineCmds
	if container {
		prefix = containerDir
		cmds = containerCmds
	}

	rootDir := debugFolder
	if outpath != "" {
		rootDir = outpath
	}
	exp := []string{}
	for _, cmd := range cmds {
		exp = append(exp, filepath.Join(rootDir, prefix, id, cmd.name))
	}

	return exp
}
