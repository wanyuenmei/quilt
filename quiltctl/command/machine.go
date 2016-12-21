package command

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
	"github.com/NetSys/quilt/db"
)

// Machine contains the options for querying machines.
type Machine struct {
	common       *commonFlags
	clientGetter client.Getter
}

// NewMachineCommand creates a new Machine command instance.
func NewMachineCommand() *Machine {
	return &Machine{
		common:       &commonFlags{},
		clientGetter: getter.New(),
	}
}

// InstallFlags sets up parsing for command line flags
func (mCmd *Machine) InstallFlags(flags *flag.FlagSet) {
	mCmd.common.InstallFlags(flags)
	flags.Usage = func() {
		fmt.Println("usage: quilt machine [-H=<daemon_host>]")
		fmt.Println("`machine` displays the status of quilt-managed machines.")
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the machine command.
func (mCmd *Machine) Parse(args []string) error {
	return nil
}

// Run retrieves and prints the requested machines.
func (mCmd *Machine) Run() int {
	c, err := mCmd.clientGetter.Client(mCmd.common.host)
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
