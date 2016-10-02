package command

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/NetSys/quilt/api"
	"github.com/NetSys/quilt/api/client"
	"github.com/NetSys/quilt/db"
)

// SubCommand defines the conversion between the user CLI flags and
// functionality within the code.
type SubCommand interface {
	// The function to run once the flags have been parsed. The return value
	// is the exit code.
	Run() int

	// Give the command line arguments to the subcommand so that it can parse
	// it for later execution.
	Parse(args []string) error

	// Print out the usage of the SubCommand.
	Usage()
}

// Stored in a variable so we can mock it out for the unit tests.
var getClient = func(host string) (client.Client, error) {
	c, err := client.New(host)
	if err != nil {
		return nil, DaemonConnectError{
			host:         host,
			connectError: err,
		}
	}
	return c, nil
}

// Get a client connected to the lead minion.
func getLeaderClient(localClient client.Client) (client.Client, error) {
	machines, err := localClient.QueryMachines()
	if err != nil {
		return nil, fmt.Errorf("unable to query machines: %s", err.Error())
	}

	// Try to figure out the lead minion's IP by asking each of the machines
	// tracked by the local daemon.
	for _, m := range machines {
		if m.PublicIP == "" {
			continue
		}

		ip, err := getLeaderIP(localClient, m.PublicIP)
		if err == nil {
			return getClient(api.RemoteAddress(ip))
		}
		log.WithError(err).Debug("Unable to get leader IP.")
	}

	return nil, errors.New("no leader found")
}

// Get the public IP of the lead minion by querying the remote machine's etcd
// table for the private IP, and then searching for the public IP in the local
// daemon.
func getLeaderIP(localClient client.Client, daemonIP string) (string, error) {
	remoteClient, err := getClient(api.RemoteAddress(daemonIP))
	if err != nil {
		return "", err
	}
	defer remoteClient.Close()

	etcds, err := remoteClient.QueryEtcd()
	if err != nil {
		return "", err
	}

	if len(etcds) == 0 || etcds[0].LeaderIP == "" {
		return "", fmt.Errorf("no leader information on host %s", daemonIP)
	}

	return getPublicIP(localClient, etcds[0].LeaderIP)
}

// Get the public IP of a machine given its private IP.
func getPublicIP(c client.Client, privateIP string) (string, error) {
	machines, err := c.QueryMachines()
	if err != nil {
		return "", err
	}

	for _, m := range machines {
		if m.PrivateIP == privateIP {
			return m.PublicIP, nil
		}
	}

	return "", fmt.Errorf("no machine with private IP %s", privateIP)
}

// Get a client connected to the local daemon and the daemon on the lead minion.
func getClients(host string) (client.Client, client.Client, error) {
	localClient, err := getClient(host)
	if err != nil {
		return nil, nil, err
	}

	leaderClient, err := getLeaderClient(localClient)
	if err != nil {
		localClient.Close()
		return nil, nil, err
	}

	return localClient, leaderClient, err
}

// Get the public IP of the machine the container is running on.
func getContainerHost(localClient client.Client, leaderClient client.Client,
	stitchID int) (string, error) {

	container, err := getContainer(leaderClient, stitchID)
	if err != nil {
		return "", err
	}

	if container.Minion == "" {
		return "", errors.New("container hasn't been scheduled yet")
	}

	return getPublicIP(localClient, container.Minion)
}

// Get the container information for the given stitchID stored by the client.
func getContainer(c client.Client, stitchID int) (db.Container, error) {
	containers, err := c.QueryContainers()
	if err != nil {
		return db.Container{}, err
	}

	for _, c := range containers {
		if c.StitchID == stitchID {
			return c, nil
		}
	}

	return db.Container{}, fmt.Errorf("no container with stitchID %d", stitchID)
}
