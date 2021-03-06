package command

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/quilt/quilt/stitch"
)

// Stop contains the options for stopping namespaces.
type Stop struct {
	namespace      string
	onlyContainers bool

	connectionHelper
}

// NewStopCommand creates a new Stop command instance.
func NewStopCommand() *Stop {
	return &Stop{}
}

// InstallFlags sets up parsing for command line flags.
func (sCmd *Stop) InstallFlags(flags *flag.FlagSet) {
	sCmd.connectionHelper.InstallFlags(flags)

	flags.StringVar(&sCmd.namespace, "namespace", "",
		"the namespace to stop")
	flags.BoolVar(&sCmd.onlyContainers, "containers", false,
		"only destroy containers")

	flags.Usage = func() {
		fmt.Println("usage: quilt stop [-H=<daemon_host>] " +
			"[-containers] [-namespace=<namespace>] <namespace>]")
		fmt.Println("`stop` creates an empty Stitch for the given namespace, " +
			"and sends it to the Quilt daemon to be executed. If no " +
			"namespace is specified, `stop` attempts to use the namespace " +
			"currently tracked by the daemon.")
		fmt.Println("The result is that resources associated with the " +
			"namespace, such as VMs, are freed.")
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the stop command.
func (sCmd *Stop) Parse(args []string) error {
	if len(args) > 0 {
		sCmd.namespace = args[0]
	}

	return nil
}

// Run stops the given namespace.
func (sCmd *Stop) Run() int {
	newCluster := stitch.Stitch{
		Namespace: sCmd.namespace,
	}
	if sCmd.namespace == "" || sCmd.onlyContainers {
		currDepl, err := getCurrentDeployment(sCmd.client)
		if err != nil {
			log.WithError(err).
				Error("Failed to get current cluster")
			return 1
		}
		if sCmd.namespace == "" {
			newCluster.Namespace = currDepl.Namespace
		}
		if sCmd.onlyContainers {
			if newCluster.Namespace != currDepl.Namespace {
				log.Error("Stopping only containers for a namespace " +
					"not tracked by the remote daemon is not " +
					"currently supported")
				return 1
			}
			newCluster.Machines = currDepl.Machines
		}
	}

	if err := sCmd.client.Deploy(newCluster.String()); err != nil {
		log.WithError(err).Error("Unable to stop namespace.")
		return 1
	}

	log.WithField("namespace", sCmd.namespace).Debug("Stopping namespace")
	return 0
}
