package command

import (
	"errors"
	"testing"

	"github.com/spf13/afero"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/util"
)

func TestMachineFlags(t *testing.T) {
	t.Parallel()

	expHost := "IP"

	machineCmd := Machine{}
	err := machineCmd.Parse([]string{"-H", expHost})

	if err != nil {
		t.Errorf("Unexpected error when parsing container args: %s", err.Error())
		return
	}

	if machineCmd.host != expHost {
		t.Errorf("Expected machine command to parse arg %s, but got %s",
			expHost, machineCmd.host)
	}
}

func TestMachineOutput(t *testing.T) {
	t.Parallel()

	res := machinesStr([]db.Machine{{ID: 1, Role: db.Master, PublicIP: "8.8.8.8"}})
	exp := "Machine-1{Role=Master, PublicIP=8.8.8.8, Connected=false}\n"
	if res != exp {
		t.Errorf("Expected machine command to print %s, but got %s.", exp, res)
	}
}

func TestContainerFlags(t *testing.T) {
	t.Parallel()

	expHost := "IP"

	containerCmd := Container{}
	err := containerCmd.Parse([]string{"-H", expHost})

	if err != nil {
		t.Errorf("Unexpected error when parsing container args: %s", err.Error())
		return
	}

	if containerCmd.host != expHost {
		t.Errorf("Expected container command to parse arg %s, but got %s",
			expHost, containerCmd.host)
	}
}

func TestContainerOutput(t *testing.T) {
	t.Parallel()

	res := containersStr([]db.Container{{ID: 1, Command: []string{"cmd", "arg"}}})
	exp := "Container-1{run  cmd arg}\n"
	if res != exp {
		t.Errorf("Expected container command to print %s, but got %s.", exp, res)
	}
}

func checkGetParsing(t *testing.T, args []string, expImport string, expErr error) {
	getCmd := Get{}
	err := getCmd.Parse(args)

	if expErr != nil {
		if err.Error() != expErr.Error() {
			t.Errorf("Expected error %s, but got %s",
				expErr.Error(), err.Error())
		}
		return
	}

	if err != nil {
		t.Errorf("Unexpected error when parsing get args: %s", err.Error())
		return
	}

	if getCmd.importPath != expImport {
		t.Errorf("Expected get command to parse arg %s, but got %s",
			expImport, getCmd.importPath)
	}
}

func TestGetFlags(t *testing.T) {
	t.Parallel()

	expImport := "spec"
	checkGetParsing(t, []string{"-import", expImport}, expImport, nil)
	checkGetParsing(t, []string{expImport}, expImport, nil)
	checkGetParsing(t, []string{}, "", errors.New("no import specified"))
}

func checkRunParsing(t *testing.T, args []string, expStitch string, expErr error) {
	runCmd := Run{}
	err := runCmd.Parse(args)

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

func TestRunFlags(t *testing.T) {
	t.Parallel()

	expStitch := "spec"
	checkRunParsing(t, []string{"-stitch", expStitch}, expStitch, nil)
	checkRunParsing(t, []string{expStitch}, expStitch, nil)
	checkRunParsing(t, []string{}, "", errors.New("no spec specified"))
}

func checkStopParsing(t *testing.T, args []string, expNamespace string, expErr error) {
	stopCmd := Stop{}
	err := stopCmd.Parse(args)

	if expErr != nil {
		if err.Error() != expErr.Error() {
			t.Errorf("Expected error %s, but got %s",
				expErr.Error(), err.Error())
		}
		return
	}

	if err != nil {
		t.Errorf("Unexpected error when parsing stop args: %s", err.Error())
		return
	}

	if stopCmd.namespace != expNamespace {
		t.Errorf("Expected stop command to parse arg %s, but got %s",
			expNamespace, stopCmd.namespace)
	}
}

func TestStopFlags(t *testing.T) {
	t.Parallel()

	expNamespace := "namespace"
	checkStopParsing(t, []string{"-namespace", expNamespace}, expNamespace, nil)
	checkStopParsing(t, []string{expNamespace}, expNamespace, nil)
}

type mockClient struct {
	runStitchArg string
}

func (c *mockClient) QueryMachines() ([]db.Machine, error) {
	return nil, nil
}

func (c *mockClient) QueryContainers() ([]db.Container, error) {
	return nil, nil
}

func (c *mockClient) Close() error {
	return nil
}

func (c *mockClient) RunStitch(stitch string) error {
	c.runStitchArg = stitch
	return nil
}

func TestStopNamespace(t *testing.T) {
	c := &mockClient{}
	getClient = func(host string) (client.Client, error) {
		return c, nil
	}

	stopCmd := &Stop{namespace: "namespace"}
	stopCmd.Run()
	expStitch := `(define AdminACL (list)) (define Namespace "namespace")`
	if c.runStitchArg != expStitch {
		t.Error("stop command invoked Quilt with the wrong stitch")
	}

	stopCmd = &Stop{}
	stopCmd.Run()
	expStitch = "(define AdminACL (list))"
	if c.runStitchArg != expStitch {
		t.Error("stop command invoked Quilt with the wrong stitch")
	}
}

func TestRunSpec(t *testing.T) {
	c := &mockClient{}
	getClient = func(host string) (client.Client, error) {
		return c, nil
	}

	stitchPath := "test.spec"
	expStitch := `(docker "nginx")`
	util.AppFs = afero.NewMemMapFs()
	util.WriteFile(stitchPath, []byte(expStitch), 0644)

	runCmd := &Run{stitch: stitchPath}
	runCmd.Run()
	if c.runStitchArg != expStitch {
		t.Error("run command invoked Quilt with the wrong stitch")
	}
}
