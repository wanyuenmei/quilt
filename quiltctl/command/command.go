package command

import (
	"flag"
)

// SubCommand defines the conversion between the user CLI flags and
// functionality within the code.
type SubCommand interface {
	// InstallFlags sets up parsing for command line flags.
	InstallFlags(*flag.FlagSet)

	// BeforeRun is called after command line flags have been parsed, but
	// before the Run method is called. It gives commands an opportunity to
	// transform parsed flags into a more consumable form.
	BeforeRun() error

	// AfterRun is called after the Run method so that commands can perform
	// post-run cleanup.
	AfterRun() error

	// The function to run once the flags have been parsed. The return value
	// is the exit code.
	Run() int

	// Give the non-flag command line arguments to the subcommand so that it can
	// parse it for later execution.
	Parse(args []string) error
}
