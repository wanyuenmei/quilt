package command

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
)

// The default namespace to use when stopping. This should match the default
// namespace used in `stitch/bindings.js`.
const defaultNamespace = "default-namespace"

// Stop contains the options for stopping namespaces.
type Stop struct {
	namespace string

	common       *commonFlags
	clientGetter client.Getter
}

// NewStopCommand creates a new Stop command instance.
func NewStopCommand() *Stop {
	return &Stop{
		clientGetter: getter.New(),
		common:       &commonFlags{},
	}
}

// InstallFlags sets up parsing for command line flags.
func (sCmd *Stop) InstallFlags(flags *flag.FlagSet) {
	sCmd.common.InstallFlags(flags)

	flags.StringVar(&sCmd.namespace, "namespace", defaultNamespace,
		"the namespace to stop")

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
	if len(args) > 0 {
		sCmd.namespace = args[0]
	}

	return nil
}

// Run stops the given namespace.
func (sCmd *Stop) Run() int {
	c, err := sCmd.clientGetter.Client(sCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer c.Close()

	specStr := fmt.Sprintf(`{"namespace": %q}`, sCmd.namespace)
	if err = c.Deploy(specStr); err != nil {
		log.WithError(err).Error("Unable to stop namespace.")
		return 1
	}

	log.WithField("namespace", sCmd.namespace).Debug("Stopping namespace")
	return 0
}
