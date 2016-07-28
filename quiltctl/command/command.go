package command

import (
	"github.com/NetSys/quilt/api/client"
)

// SubCommand defines the conversion between the user CLI flags and
// functionality within the code.
type SubCommand interface {
	// The function to run once the flags have been parsed. The return value
	// is the exit code.
	Run() int

	// Give the command line arguments to the subcommand so that it can parse
	// it for later execution.
	Parse(args []string) error

	// Print out the usage of the SubCommand.
	Usage()
}

// Stored in a variable so we can mock it out for the unit tests.
var getClient = func(host string) (client.Client, error) {
	c, err := client.New(host)
	if err != nil {
		return nil, DaemonConnectError{
			host:         host,
			connectError: err,
		}
	}
	return c, nil
}
