package command

import (
	"flag"
	"fmt"

	"github.com/quilt/quilt/version"
)

// Version prints the Quilt version information.
type Version struct{}

var versionUsage = `usage: quilt version
Show the Quilt version information.`

// InstallFlags sets up parsing for command line flags.
func (vCmd *Version) InstallFlags(flags *flag.FlagSet) {
	flags.Usage = func() {
		fmt.Println(versionUsage)
	}
}

// Parse parses the command line arguments for the version command.
func (vCmd *Version) Parse(args []string) error {
	return nil
}

// Run prints the version information.
func (vCmd *Version) Run() int {
	fmt.Println(version.Version)
	return 0
}
