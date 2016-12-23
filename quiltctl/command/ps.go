package command

import (
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
)

// Ps contains the options for querying machines and containers.
type Ps struct {
	common       *commonFlags
	clientGetter client.Getter
}

// NewPsCommand creates a new Ps command instance.
func NewPsCommand() *Ps {
	return &Ps{
		common:       &commonFlags{},
		clientGetter: getter.New(),
	}
}

// InstallFlags sets up parsing for command line flags
func (pCmd *Ps) InstallFlags(flags *flag.FlagSet) {
	pCmd.common.InstallFlags(flags)
	flags.Usage = func() {
		fmt.Println("usage: quilt ps [-H=<daemon_host>]")
		fmt.Println("`ps` displays the status of quilt-managed " +
			"machines and containers.")

		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the ps command.
func (pCmd *Ps) Parse(args []string) error {
	return nil
}

// Run retrieves and prints all machines and containers.
func (pCmd *Ps) Run() int {
	localClient, err := pCmd.clientGetter.Client(pCmd.common.host)
	if err != nil {
		log.WithError(err).Error("Error connecting to local client.")
		return 1
	}
	defer localClient.Close()

	machines, err := localClient.QueryMachines()
	if err != nil {
		log.WithError(err).Error("Unable to query machines.")
		return 1
	}

	fmt.Println("MACHINES")
	writeMachines(os.Stdout, machines)

	leaderClient, err := pCmd.clientGetter.LeaderClient(localClient)
	if err != nil {
		log.WithError(err).Error("Error connecting to leader.")
		return 1
	}
	defer leaderClient.Close()

	connections, err := leaderClient.QueryConnections()
	if err != nil {
		log.WithError(err).Error("Unable to query connections.")
	}

	containers, err := leaderClient.QueryContainers()
	if err != nil {
		log.WithError(err).Error("Unable to query containers.")
		return 1
	}

	fmt.Println()
	fmt.Println("CONTAINERS")
	writeContainers(os.Stdout, containers, machines, connections)
	return 0
}
