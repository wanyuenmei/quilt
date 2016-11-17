package command

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	log "github.com/Sirupsen/logrus"
	logrusTestHook "github.com/Sirupsen/logrus/hooks/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	clientMock "github.com/NetSys/quilt/api/client/mocks"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/quiltctl/testutils"
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
		util.AppFs = afero.NewMemMapFs()

		mockGetter := new(testutils.Getter)
		c := &clientMock.Client{}
		mockGetter.On("Client", mock.Anything).Return(c, nil)

		logHook := logrusTestHook.NewGlobal()

		util.WriteFile(test.file.path, []byte(test.file.contents), 0644)
		runCmd := NewRunCommand()
		runCmd.clientGetter = mockGetter
		runCmd.stitch = test.path
		exitCode := runCmd.Run()

		assert.Equal(t, test.expExitCode, exitCode)
		assert.Equal(t, test.expDeployArg, c.DeployArg)

		assert.Equal(t, len(test.expEntries), len(logHook.Entries))
		for i, entry := range logHook.Entries {
			assert.Equal(t, test.expEntries[i].Message, entry.Message)
			assert.Equal(t, test.expEntries[i].Level, entry.Level)
		}
	}
}

type diffTest struct {
	curr, new, exp string
}

func TestDeploymentDiff(t *testing.T) {
	t.Parallel()

	tests := []diffTest{
		{
			curr: "{}",
			new:  "{}",
			exp:  "",
		},
		{
			curr: `{"Machines":[{"Provider":"Amazon"}]}`,
			new:  `{"Machines":[]}`,
			exp: `--- Current
+++ Proposed
@@ -1,7 +1,3 @@
 {
-	"Machines": [
-		{
-			"Provider": "Amazon"
-		}
-	]
+	"Machines": []
 }
`,
		},
		{
			curr: `{"Machines":[{"Provider":"Amazon"},` +
				`{"Provider":"Google"}]}`,
			new: `{"Machines":[{"Provider":"Google"}]}`,
			exp: `--- Current
+++ Proposed
@@ -1,8 +1,5 @@
 {
 	"Machines": [
-		{
-			"Provider": "Amazon"
-		},
 		{
 			"Provider": "Google"
 		}
`,
		},
		{
			curr: `{"Machines":[{"Provider":"Amazon"},` +
				`{"Provider":"Google"}]}`,
			new: `{"Machines":[{"Provider":"Vagrant"}]}`,
			exp: `--- Current
+++ Proposed
@@ -1,10 +1,7 @@
 {
 	"Machines": [
 		{
-			"Provider": "Amazon"
-		},
-		{
-			"Provider": "Google"
+			"Provider": "Vagrant"
 		}
 	]
 }
`,
		},
	}

	for _, test := range tests {
		diff, err := diffDeployment(test.curr, test.new)
		assert.Nil(t, err)
		assert.Equal(t, test.exp, diff)
	}
}

type confirmTest struct {
	inputs []string
	exp    bool
}

func TestConfirm(t *testing.T) {
	tests := []confirmTest{
		{
			inputs: []string{"y"},
			exp:    true,
		},
		{
			inputs: []string{"yes"},
			exp:    true,
		},
		{
			inputs: []string{"YES"},
			exp:    true,
		},
		{
			inputs: []string{"n"},
			exp:    false,
		},
		{
			inputs: []string{"no"},
			exp:    false,
		},
		{
			inputs: []string{"foo", "no"},
			exp:    false,
		},
		{
			inputs: []string{"foo", "no", "yes"},
			exp:    false,
		},
	}
	for _, test := range tests {
		in := bytes.NewBufferString(strings.Join(test.inputs, "\n"))
		res, err := confirm(in, "")
		assert.Nil(t, err)
		assert.Equal(t, test.exp, res)
	}
}

func TestPromptsUser(t *testing.T) {
	oldConfirm := confirm
	defer func() {
		confirm = oldConfirm
	}()

	util.AppFs = afero.NewMemMapFs()
	for _, confirmResp := range []bool{true, false} {
		confirm = func(in io.Reader, prompt string) (bool, error) {
			return confirmResp, nil
		}

		mockGetter := new(testutils.Getter)
		c := &clientMock.Client{
			ClusterReturn: []db.Cluster{
				{
					Spec: `{"old":"spec"}`,
				},
			},
		}
		mockGetter.On("Client", mock.Anything).Return(c, nil)

		util.WriteFile("test.js", []byte(""), 0644)
		runCmd := NewRunCommand()
		runCmd.clientGetter = mockGetter
		runCmd.stitch = "test.js"
		runCmd.Run()
		assert.Equal(t, confirmResp, c.DeployArg != "")
	}
}

func TestRunFlags(t *testing.T) {
	t.Parallel()

	expStitch := "spec"
	checkRunParsing(t, []string{"-stitch", expStitch}, Run{stitch: expStitch}, nil)
	checkRunParsing(t, []string{expStitch}, Run{stitch: expStitch}, nil)
	checkRunParsing(t, []string{"-f", expStitch},
		Run{force: true, stitch: expStitch}, nil)
	checkRunParsing(t, []string{}, Run{}, errors.New("no spec specified"))
}

func checkRunParsing(t *testing.T, args []string, expFlags Run, expErr error) {
	runCmd := NewRunCommand()
	err := parseHelper(runCmd, args)

	if expErr != nil {
		if err.Error() != expErr.Error() {
			t.Errorf("Expected error %s, but got %s",
				expErr.Error(), err.Error())
		}
		return
	}

	assert.Nil(t, err)
	assert.Equal(t, expFlags.stitch, runCmd.stitch)
	assert.Equal(t, expFlags.force, runCmd.force)
}
