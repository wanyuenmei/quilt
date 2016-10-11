package command

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/db"
)

// Machine contains the options for querying machines.
type Machine struct {
	*commonFlags
}

// NewMachineCommand creates a new Machine command instance.
func NewMachineCommand() *Machine {
	return &Machine{
		commonFlags: &commonFlags{},
	}
}

// Parse parses the command line arguments for the machine command.
func (mCmd *Machine) Parse(args []string) error {
	return nil
}

// Run retrieves and prints the requested machines.
func (mCmd *Machine) Run() int {
	c, err := getClient(mCmd.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer c.Close()

	machines, err := c.QueryMachines()
	if err != nil {
		log.WithError(err).Error("Unable to query machines.")
		return 1
	}

	str := machinesStr(machines)
	fmt.Print(str)

	return 0
}

func machinesStr(machines []db.Machine) string {
	var machinesStr string
	for _, m := range db.SortMachines(machines) {
		machinesStr += fmt.Sprintf("%v\n", m)
	}

	return machinesStr
}
