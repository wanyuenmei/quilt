package command

import (
	"testing"

	"github.com/NetSys/quilt/db"
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
