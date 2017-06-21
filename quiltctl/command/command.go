package command

import (
	"flag"
)

// SubCommand defines the conversion between the user CLI flags and
// functionality within the code.
type SubCommand interface {
	// InstallFlags sets up parsing for command line flags.
	InstallFlags(*flag.FlagSet)

	// The function to run once the flags have been parsed. The return value
	// is the exit code.
	Run() int

	// Give the non-flag command line arguments to the subcommand so that it can
	// parse it for later execution.
	Parse(args []string) error
}
