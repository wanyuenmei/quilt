// Package network manages the network services of the application dataplane.  This
// means ensuring that containers can find and communicate with each other in accordance
// with the policy specification.  It achieves this by manipulating IP addresses and
// hostnames within the containers, Open vSwitch on each running worker, and the OVN
// controller.
package network

import (
	"github.com/quilt/quilt/db"
	"github.com/quilt/quilt/join"
	"github.com/quilt/quilt/minion/ipdef"
	"github.com/quilt/quilt/minion/ovsdb"

	log "github.com/Sirupsen/logrus"
)

const lSwitch = "quilt"

// Run blocks implementing the network services.
func Run(conn db.Conn) {
	go runNat(conn)
	go runDNS(conn)
	go runUpdateIPs(conn)

	for range conn.TriggerTick(30, db.MinionTable, db.ContainerTable,
		db.ConnectionTable, db.LabelTable, db.EtcdTable).C {
		if conn.EtcdLeader() {
			runMaster(conn)
		}
	}
}

// The leader of the cluster is responsible for properly configuring OVN northd for
// container networking.  This simply means creating a logical port for each container
// and label.  The specialized OpenFlow rules Quilt requires are managed by the workers
// individuallly.
func runMaster(conn db.Conn) {
	var labels []db.Label
	var containers []db.Container
	var connections []db.Connection
	conn.Txn(db.ConnectionTable, db.ContainerTable, db.EtcdTable,
		db.LabelTable, db.MinionTable).Run(func(view db.Database) error {

		labels = view.SelectFromLabel(func(label db.Label) bool {
			return label.IP != ""
		})

		containers = view.SelectFromContainer(func(dbc db.Container) bool {
			return dbc.IP != ""
		})

		connections = view.SelectFromConnection(nil)
		return nil
	})

	ovsdbClient, err := ovsdb.Open()
	if err != nil {
		log.WithError(err).Error("Failed to connect to OVSDB.")
		return
	}
	defer ovsdbClient.Disconnect()

	switchExists, err := ovsdbClient.LogicalSwitchExists(lSwitch)
	if err != nil {
		log.WithError(err).Error("Failed to check existence of logical switch")
		return
	}

	if !switchExists {
		if err := ovsdbClient.CreateLogicalSwitch(lSwitch); err != nil {
			log.WithError(err).Error("Failed to create logical switch")
			return
		}
	}

	lports, err := ovsdbClient.ListSwitchPorts()
	if err != nil {
		log.WithError(err).Error("Failed to list OVN switch ports.")
		return
	}

	dbcKey := func(val interface{}) interface{} {
		return val.(db.Container).IP
	}
	portKey := func(val interface{}) interface{} {
		return val.(ovsdb.SwitchPort).Name
	}

	_, ovsps, dbcs := join.HashJoin(ovsdb.SwitchPortSlice(lports),
		db.ContainerSlice(containers), portKey, dbcKey)

	for _, dbcIface := range dbcs {
		dbc := dbcIface.(db.Container)
		lport := ovsdb.SwitchPort{
			Name: dbc.IP,
			// OVN represents network interfaces with the empty string.
			Type:      "",
			Addresses: []string{ipdef.IPStrToMac(dbc.IP) + " " + dbc.IP},
		}
		err := ovsdbClient.CreateSwitchPort(lSwitch, lport)
		if err != nil {
			log.WithError(err).Warnf(
				"Failed to create logical switch port: %s", dbc.IP)
		} else {
			log.Infof("New logical switch port: %s", dbc.IP)
		}
	}

	for _, ovsp := range ovsps {
		lport := ovsp.(ovsdb.SwitchPort)
		if err := ovsdbClient.DeleteSwitchPort(lSwitch, lport); err != nil {
			log.WithError(err).Warnf(
				"Failed to delete logical switch port: %s", lport.Name)
		} else {
			log.Infof("Delete logical switch port: %s", lport.Name)
		}
	}

	updateACLs(ovsdbClient, connections, labels)
}
