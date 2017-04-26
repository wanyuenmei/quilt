package command

import (
	"flag"
	"fmt"

	"github.com/quilt/quilt/api/client"
	"github.com/quilt/quilt/api/client/getter"
	"github.com/quilt/quilt/version"

	log "github.com/Sirupsen/logrus"
)

// Version prints the Quilt version information.
type Version struct {
	common       *commonFlags
	clientGetter client.Getter
}

// NewVersionCommand creates a new Version command instance.
func NewVersionCommand() *Version {
	return &Version{
		common:       &commonFlags{},
		clientGetter: getter.New(),
	}
}

var versionUsage = `usage: quilt version [-H=<daemon_host>]
Show the Quilt version information.`

// InstallFlags sets up parsing for command line flags.
func (vCmd *Version) InstallFlags(flags *flag.FlagSet) {
	vCmd.common.InstallFlags(flags)
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
	fmt.Println("Client:", version.Version)

	daemonVersion, err := vCmd.getDaemonVersion()
	if err != nil {
		log.WithError(err).Error("Failed to get daemon version")
		return 1
	}
	fmt.Println("Daemon:", daemonVersion)

	return 0
}

func (vCmd Version) getDaemonVersion() (string, error) {
	client, err := vCmd.clientGetter.Client(vCmd.common.host)
	if err != nil {
		return "", err
	}
	defer client.Close()

	version, err := client.Version()
	if err != nil {
		return "", err
	}

	return version, nil
}
