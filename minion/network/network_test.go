package network

import (
	"fmt"
	"testing"

	"github.com/NetSys/quilt/db"
	"github.com/NetSys/quilt/join"
	"github.com/NetSys/quilt/minion/ovsdb"
)

func TestACL(t *testing.T) {
	client := ovsdb.NewFakeOvsdbClient()
	client.CreateLogicalSwitch(lSwitch)
	ovsdb.Open = func() (ovsdb.Client, error) {
		return client, nil
	}

	defaultAclCores := []ovsdb.AclCore{
		{
			Priority:  0,
			Match:     "ip",
			Action:    "drop",
			Direction: "to-lport",
		},
		{
			Priority:  0,
			Match:     "ip",
			Action:    "drop",
			Direction: "from-lport",
		},
	}

	// `expACLs` should NOT contain the default rules.
	checkAcl := func(connections []db.Connection, labels []db.Label,
		containers []db.Container, expectedAclCores []ovsdb.AclCore,
		resetClient bool) {
		if resetClient {
			client = ovsdb.NewFakeOvsdbClient()
			client.CreateLogicalSwitch(lSwitch)
		}
		updateACLs(connections, labels, containers)

		res, err := client.ListACLs(lSwitch)
		if err != nil {
			t.Error(err)
		}
		updateACLs(connections, labels, containers)

		res, err = client.ListACLs(lSwitch)
		if err != nil {
			t.Error(err)
		}

		expectedAclCores = append(defaultAclCores, expectedAclCores...)

		key := func(val interface{}) interface{} {
			switch v := val.(type) {
			case ovsdb.AclCore:
				return v
			case ovsdb.Acl:
				return v.Core
			}
			return nil
		}

		pair, _, _ := join.HashJoin(AclCoreSlice(expectedAclCores),
			AclSlice(res), key, key)
		if len(pair) != len(expectedAclCores) {
			t.Error("Local ACLs do not match ovsdbACLs.")
		}
	}

	redLabelIP := "8.8.8.8"
	blueLabelIP := "9.9.9.9"
	yellowLabelIP := "10.10.10.10"
	purpleLabelIP := "12.12.12.12"
	redBlueLabelIP := "13.13.13.13"
	redLabel := db.Label{Label: "red",
		IP: redLabelIP}
	blueLabel := db.Label{Label: "blue",
		IP: blueLabelIP}
	yellowLabel := db.Label{Label: "yellow",
		IP: yellowLabelIP}
	purpleLabel := db.Label{Label: "purple",
		IP: purpleLabelIP}
	redBlueLabel := db.Label{Label: "redBlue",
		IP: redBlueLabelIP}
	allLabels := []db.Label{redLabel, blueLabel, yellowLabel, purpleLabel,
		redBlueLabel}

	redContainerIP := "100.1.1.1"
	blueContainerIP := "100.1.1.2"
	yellowContainerIP := "100.1.1.3"
	purpleContainerIP := "100.1.1.4"
	redContainer := db.Container{IP: redContainerIP,
		Labels: []string{"red", "redBlue"},
	}
	blueContainer := db.Container{IP: blueContainerIP,
		Labels: []string{"blue", "redBlue"},
	}
	yellowContainer := db.Container{IP: yellowContainerIP,
		Labels: []string{"yellow"},
	}
	purpleContainer := db.Container{IP: purpleContainerIP,
		Labels: []string{"purple"},
	}
	allContainers := []db.Container{redContainer, blueContainer, yellowContainer,
		purpleContainer}

	matchFmt := "ip4.src==%s && ip4.dst==%s && " +
		"(%d <= udp.dst <= %d || %[3]d <= tcp.dst <= %[4]d)"
	reverseFmt := "ip4.src==%s && ip4.dst==%s && " +
		"(%d <= udp.src <= %d || %[3]d <= tcp.src <= %[4]d)"
	icmpFmt := "ip4.src==%s && ip4.dst==%s && icmp"

	// No connections should result in no ACLs but the default drop rules.
	checkAcl([]db.Connection{}, []db.Label{}, []db.Container{}, []ovsdb.AclCore{},
		true)

	// Test one connection (with range)
	checkAcl([]db.Connection{
		{From: "red",
			To:      "blue",
			MinPort: 80,
			MaxPort: 81}},
		allLabels,
		allContainers,
		[]ovsdb.AclCore{
			{Direction: "to-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					blueLabelIP, 80, 81),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					blueLabelIP, 80, 81),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(reverseFmt, blueLabelIP,
					redContainerIP, 80, 81),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(reverseFmt, blueLabelIP,
					redContainerIP, 80, 81),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					blueLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					blueLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, blueLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, blueLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			}},
		true)

	// Test connecting from label with multiple containers
	checkAcl([]db.Connection{
		{From: "redBlue",
			To:      "yellow",
			MinPort: 80,
			MaxPort: 80}},
		allLabels,
		allContainers,
		[]ovsdb.AclCore{
			{Direction: "to-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					yellowLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					yellowLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(reverseFmt, yellowLabelIP,
					redContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(reverseFmt, yellowLabelIP,
					redContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(matchFmt, blueContainerIP,
					yellowLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(matchFmt, blueContainerIP,
					yellowLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(reverseFmt, yellowLabelIP,
					blueContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(reverseFmt, yellowLabelIP,
					blueContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					yellowLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					yellowLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, yellowLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, yellowLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, blueContainerIP,
					yellowLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, blueContainerIP,
					yellowLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, yellowLabelIP,
					blueContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, yellowLabelIP,
					blueContainerIP),
				Action:   "allow",
				Priority: 1,
			}},
		true)

	// Test removing a connection
	checkAcl([]db.Connection{
		{From: "red",
			To:      "blue",
			MinPort: 80,
			MaxPort: 80}},
		allLabels,
		allContainers,
		[]ovsdb.AclCore{
			{Direction: "to-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					blueLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					blueLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(reverseFmt, blueLabelIP,
					redContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(reverseFmt, blueLabelIP,
					redContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					blueLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					blueLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, blueLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, blueLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			}},
		true)
	checkAcl([]db.Connection{},
		allLabels,
		allContainers,
		[]ovsdb.AclCore{},
		false)

	// Test removing one connection, but not another
	checkAcl([]db.Connection{
		{From: "red",
			To:      "blue",
			MinPort: 80,
			MaxPort: 80},
		{From: "yellow",
			To:      "purple",
			MinPort: 80,
			MaxPort: 80}},
		allLabels,
		allContainers,
		[]ovsdb.AclCore{
			{Direction: "to-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					blueLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(matchFmt, redContainerIP,
					blueLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(reverseFmt, blueLabelIP,
					redContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(reverseFmt, blueLabelIP,
					redContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(matchFmt, yellowContainerIP,
					purpleLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(matchFmt, yellowContainerIP,
					purpleLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(reverseFmt, purpleLabelIP,
					yellowContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(reverseFmt, purpleLabelIP,
					yellowContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					blueLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, redContainerIP,
					blueLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, blueLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, blueLabelIP,
					redContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, yellowContainerIP,
					purpleLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, yellowContainerIP,
					purpleLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, purpleLabelIP,
					yellowContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, purpleLabelIP,
					yellowContainerIP),
				Action:   "allow",
				Priority: 1,
			}},
		true)
	checkAcl([]db.Connection{
		{From: "yellow",
			To:      "purple",
			MinPort: 80,
			MaxPort: 80}},
		allLabels,
		allContainers,
		[]ovsdb.AclCore{
			{Direction: "to-lport",
				Match: fmt.Sprintf(matchFmt, yellowContainerIP,
					purpleLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(matchFmt, yellowContainerIP,
					purpleLabelIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(reverseFmt, purpleLabelIP,
					yellowContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(reverseFmt, purpleLabelIP,
					yellowContainerIP, 80, 80),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, yellowContainerIP,
					purpleLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, yellowContainerIP,
					purpleLabelIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "to-lport",
				Match: fmt.Sprintf(icmpFmt, purpleLabelIP,
					yellowContainerIP),
				Action:   "allow",
				Priority: 1,
			},
			{Direction: "from-lport",
				Match: fmt.Sprintf(icmpFmt, purpleLabelIP,
					yellowContainerIP),
				Action:   "allow",
				Priority: 1,
			}},
		true)
}

type ACLList []ovsdb.Acl

func (lst ACLList) Len() int {
	return len(lst)
}

func (lst ACLList) Swap(i, j int) {
	lst[i], lst[j] = lst[j], lst[i]
}

func (lst ACLList) Less(i, j int) bool {
	l, r := lst[i], lst[j]

	switch {
	case l.Core.Match != r.Core.Match:
		return l.Core.Match < r.Core.Match
	case l.Core.Direction != r.Core.Direction:
		return l.Core.Direction < r.Core.Direction
	case l.Core.Action != r.Core.Action:
		return l.Core.Action < r.Core.Action
	default:
		return l.Core.Priority < r.Core.Priority
	}
}

type AclSlice []ovsdb.Acl

func (acls AclSlice) Get(i int) interface{} {
	return acls[i]
}

func (acls AclSlice) Len() int {
	return len(acls)
}

type AclCoreSlice []ovsdb.AclCore

func (cores AclCoreSlice) Get(i int) interface{} {
	return cores[i]
}

func (cores AclCoreSlice) Len() int {
	return len(cores)
}
