package client

import (
	"errors"
	"fmt"

	"github.com/quilt/quilt/api"
	"github.com/quilt/quilt/api/util"
	"github.com/quilt/quilt/db"

	log "github.com/Sirupsen/logrus"
)

// Leader obtains a Client connected to the Leader of the cluster.
func Leader(machines []db.Machine) (Client, error) {
	// Try to figure out the lead minion's IP by asking each of the machines.
	for _, m := range machines {
		if m.PublicIP == "" {
			continue
		}

		ip, err := getLeaderIP(machines, m.PublicIP)
		if err == nil {
			return newClient(api.RemoteAddress(ip))
		}
		log.WithError(err).Debug("Unable to get leader IP")
	}

	return nil, errors.New("no leader found")
}

// Get the public IP of the lead minion by querying the remote machine's etcd
// table for the private IP, and then searching for the public IP in the local
// daemon.
func getLeaderIP(machines []db.Machine, daemonIP string) (string, error) {
	remoteClient, err := newClient(api.RemoteAddress(daemonIP))
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

	return util.GetPublicIP(machines, etcds[0].LeaderIP)
}

// New is saved in a variable to facilitate injecting test clients for
// unit testing.
var newClient = New
