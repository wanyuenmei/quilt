package command

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api"
	"github.com/NetSys/quilt/db"
)

// Machine contains the options for querying machines.
type Machine struct {
	host string

	flags *flag.FlagSet
}

func (mCmd *Machine) createFlagSet() {
	flags := flag.NewFlagSet("machines", flag.ExitOnError)
	flags.StringVar(&mCmd.host, "H", api.DefaultSocket, "the host to connect to")
	mCmd.flags = flags
}

// Parse parses the command line arguments for the machine command.
func (mCmd *Machine) Parse(args []string) error {
	mCmd.createFlagSet()
	return mCmd.flags.Parse(args)
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
		log.Error(DaemonResponseError{responseError: err})
		return 1
	}

	str := machinesStr(machines)
	fmt.Print(str)

	return 0
}

func machinesStr(machines []db.Machine) string {
	var machinesStr string
	for _, m := range machines {
		machinesStr += fmt.Sprintf("%v\n", m)
	}

	return machinesStr
}

// Usage prints the usage for the machine command.
func (mCmd *Machine) Usage() {
	mCmd.Usage()
}
