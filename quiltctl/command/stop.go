package command

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

// Stop contains the options for stopping namespaces.
type Stop struct {
	namespace string

	common *commonFlags
}

// NewStopCommand creates a new Stop command instance.
func NewStopCommand() *Stop {
	return &Stop{
		common: &commonFlags{},
	}
}

// InstallFlags sets up parsing for command line flags.
func (sCmd *Stop) InstallFlags(flags *flag.FlagSet) {
	sCmd.common.InstallFlags(flags)

	flags.StringVar(&sCmd.namespace, "namespace", "", "the namespace to stop")

	flags.Usage = func() {
		fmt.Println("usage: quilt stop [-H=<daemon_host>] " +
			"[-namespace=<namespace>] <namespace>]")
		fmt.Println("`stop` creates an empty Stitch for the given namespace, " +
			"and sends it to the Quilt daemon to be executed.")
		fmt.Println("The result is that resources associated with the " +
			"namespace, such as VMs, are freed.")
		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the stop command.
func (sCmd *Stop) Parse(args []string) error {
	if sCmd.namespace == "" {
		if len(args) > 0 {
			sCmd.namespace = args[0]
		}
	}

	return nil
}

// Run stops the given namespace.
func (sCmd *Stop) Run() int {
	namespace := sCmd.namespace
	var specStr string
	if namespace != "" {
		specStr += fmt.Sprintf("createDeployment({namespace: %q});", namespace)
	}

	c, err := getClient(sCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer c.Close()

	if err = c.RunStitch(specStr); err != nil {
		log.WithError(err).Error("Unable to stop namespace.")
		return 1
	}

	fmt.Printf("Successfully began stopping namespace `%s`.\n", namespace)

	return 0
}
