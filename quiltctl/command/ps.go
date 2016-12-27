package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
	"github.com/NetSys/quilt/db"
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
	if err := pCmd.run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return 1
	}
	return 0
}

func (pCmd *Ps) run() error {
	localClient, err := pCmd.clientGetter.Client(pCmd.common.host)
	if err != nil {
		return fmt.Errorf("error connecting to quilt daemon: %s", err)
	}
	defer localClient.Close()

	var connections []db.Connection
	var containers []db.Container
	var machines []db.Machine

	connectionErr := make(chan error)
	containerErr := make(chan error)
	machineErr := make(chan error)

	go func() {
		machines, err = localClient.QueryMachines()
		machineErr <- err
	}()

	leaderClient, leadErr := pCmd.clientGetter.LeaderClient(localClient)
	if leadErr == nil {
		defer leaderClient.Close()

		go func() {
			connections, err = leaderClient.QueryConnections()
			connectionErr <- err
		}()

		go func() {
			containers, err = leaderClient.QueryContainers()
			containerErr <- err
		}()
	}

	if err := <-machineErr; err != nil {
		return fmt.Errorf("unable to query machines: %s", err)
	}

	fmt.Println("MACHINES")
	writeMachines(os.Stdout, machines)
	fmt.Println()

	if leadErr != nil {
		return fmt.Errorf("unable to connect to a cluster leader: %s", leadErr)
	}
	if err := <-connectionErr; err != nil {
		return fmt.Errorf("unable to query connections: %s", err)
	}
	if err := <-containerErr; err != nil {
		return fmt.Errorf("unable to query containers: %s", err)
	}

	fmt.Println("CONTAINERS")
	writeContainers(os.Stdout, containers, machines, connections)

	return nil
}
