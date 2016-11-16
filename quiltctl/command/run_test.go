package command

import (
	"errors"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	logrusTestHook "github.com/Sirupsen/logrus/hooks/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/util"
)

type file struct {
	path, contents string
}

type runTest struct {
	file         file
	path         string
	expExitCode  int
	expDeployArg string
	expEntries   []log.Entry
}

func TestRunSpec(t *testing.T) {
	os.Setenv("QUILT_PATH", "/quilt_path")

	exJavascript := `deployment.deploy(new Machine({}));`
	exJSON := `{"Containers":[],"Labels":[],"Connections":[],"Placements":[],` +
		`"Machines":[{"Provider":"","Role":"","Size":"",` +
		`"CPU":{"Min":0,"Max":0},"RAM":{"Min":0,"Max":0},"DiskSize":0,` +
		`"Region":"","SSHKeys":[]}],"AdminACL":[],"MaxPrice":0,` +
		`"Namespace":"default-namespace","Invariants":[]}`
	tests := []runTest{
		{
			file: file{
				path:     "test.js",
				contents: exJavascript,
			},
			path:         "test.js",
			expExitCode:  0,
			expDeployArg: exJSON,
		},
		{
			path:        "dne.js",
			expExitCode: 1,
			expEntries: []log.Entry{
				{
					Message: "open /quilt_path/dne.js: " +
						"file does not exist",
					Level: log.ErrorLevel,
				},
			},
		},
		{
			path:        "/dne.js",
			expExitCode: 1,
			expEntries: []log.Entry{
				{
					Message: "open /dne.js: file does not exist",
					Level:   log.ErrorLevel,
				},
			},
		},
		{
			file: file{
				path:     "/quilt_path/in_quilt_path.js",
				contents: exJavascript,
			},
			path:         "in_quilt_path",
			expDeployArg: exJSON,
		},
	}
	for _, test := range tests {
		c := &mockClient{}
		getClient = func(host string) (client.Client, error) {
			return c, nil
		}
		util.AppFs = afero.NewMemMapFs()

		logHook := logrusTestHook.NewGlobal()

		util.WriteFile(test.file.path, []byte(test.file.contents), 0644)
		runCmd := NewRunCommand()
		runCmd.stitch = test.path
		exitCode := runCmd.Run()

		assert.Equal(t, test.expExitCode, exitCode)
		assert.Equal(t, test.expDeployArg, c.deployArg)

		assert.Equal(t, len(test.expEntries), len(logHook.Entries))
		for i, entry := range logHook.Entries {
			assert.Equal(t, test.expEntries[i].Message, entry.Message)
			assert.Equal(t, test.expEntries[i].Level, entry.Level)
		}
	}
}

func TestRunFlags(t *testing.T) {
	t.Parallel()

	expStitch := "spec"
	checkRunParsing(t, []string{"-stitch", expStitch}, expStitch, nil)
	checkRunParsing(t, []string{expStitch}, expStitch, nil)
	checkRunParsing(t, []string{}, "", errors.New("no spec specified"))
}

func checkRunParsing(t *testing.T, args []string, expStitch string, expErr error) {
	runCmd := NewRunCommand()
	err := parseHelper(runCmd, args)

	if expErr != nil {
		if err.Error() != expErr.Error() {
			t.Errorf("Expected error %s, but got %s",
				expErr.Error(), err.Error())
		}
		return
	}

	if err != nil {
		t.Errorf("Unexpected error when parsing run args: %s", err.Error())
		return
	}

	if runCmd.stitch != expStitch {
		t.Errorf("Expected run command to parse arg %s, but got %s",
			expStitch, runCmd.stitch)
	}
}
