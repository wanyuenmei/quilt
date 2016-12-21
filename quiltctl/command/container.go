package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/api/client/getter"
	"github.com/NetSys/quilt/db"
)

// Container contains the options for querying containers.
type Container struct {
	common       *commonFlags
	clientGetter client.Getter
}

// NewContainerCommand creates a new Container command instance.
func NewContainerCommand() *Container {
	return &Container{
		clientGetter: getter.New(),
		common:       &commonFlags{},
	}
}

// InstallFlags sets up parsing for command line flags
func (cCmd *Container) InstallFlags(flags *flag.FlagSet) {
	cCmd.common.InstallFlags(flags)
	flags.Usage = func() {
		fmt.Println("usage: quilt container [-H=<daemon_host>]")
		fmt.Println("`container` displays the status of quilt-managed " +
			"Docker containers.")

		flags.PrintDefaults()
	}
}

// Parse parses the command line arguments for the container command.
func (cCmd *Container) Parse(args []string) error {
	return nil
}

// Run retrieves and prints the requested containers.
func (cCmd *Container) Run() int {
	localClient, err := cCmd.clientGetter.Client(cCmd.common.host)
	if err != nil {
		log.Error(err)
		return 1
	}
	defer localClient.Close()

	c, err := cCmd.clientGetter.LeaderClient(localClient)
	if err != nil {
		log.WithError(err).Error("Error connecting to leader.")
		return 1
	}
	defer c.Close()

	containers, err := c.QueryContainers()
	if err != nil {
		log.WithError(err).Error("Unable to query containers.")
		return 1
	}

	machines, err := localClient.QueryMachines()
	if err != nil {
		log.WithError(err).Error("Unable to query machines")
		return 1
	}

	writeContainers(os.Stdout, machines, containers)
	return 0
}

func writeContainers(fd io.Writer, machines []db.Machine, containers []db.Container) {
	w := tabwriter.NewWriter(fd, 0, 0, 4, ' ', 0)
	defer w.Flush()
	fmt.Fprintln(w, "STITCH ID\tCONTAINER\tMACHINE\tIMAGE\tCOMMAND\tLABELS")

	ipIDMap := map[string]int{}
	for _, m := range machines {
		ipIDMap[m.PrivateIP] = m.ID
	}

	machineDBC := map[int][]db.Container{}
	for _, dbc := range containers {
		id := ipIDMap[dbc.Minion]
		machineDBC[id] = append(machineDBC[id], dbc)
	}

	var machineIDs []int
	for key := range machineDBC {
		machineIDs = append(machineIDs, key)
	}
	sort.Ints(machineIDs)

	for _, machineID := range machineIDs {
		machineStr := ""
		if machineID != 0 {
			machineStr = fmt.Sprintf("Machine-%d", machineID)
		}

		for _, dbc := range db.SortContainers(machineDBC[machineID]) {
			cmd := strings.Join(dbc.Command, " ")
			labels := strings.Join(dbc.Labels, ", ")
			fmt.Fprintf(w, "%v\tContainer-%v\t%v\t%v\t\"%v\"\t%v\n",
				dbc.StitchID, dbc.ID, machineStr, dbc.Image, cmd, labels)
		}
	}
}
