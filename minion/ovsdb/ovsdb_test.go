package ovsdb

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/NetSys/quilt/join"
	ovs "github.com/socketplane/libovsdb"
)

func TestCreateLogicalSwitch(t *testing.T) {
	ovsdbClient := NewFakeOvsdbClient()

	// Create new switch.
	lswitch := "test-switch"
	if err := ovsdbClient.CreateLogicalSwitch(lswitch); err != nil {
		t.Error(err)
	}

	// Check existence of created switch.
	switchReply, err := ovsdbClient.transact("OVN_Northbound", ovs.Operation{
		Op:    "select",
		Table: "Logical_Switch",
		Where: newCondition("name", "==", lswitch),
	})
	if err != nil {
		t.Error(err)
	}
	if len(switchReply[0].Rows) != 1 {
		t.Error("logical switch creation failed")
	}

	// Try to create the same switch. Should now return error since it exists.
	if err := ovsdbClient.CreateLogicalSwitch(lswitch); err == nil {
		t.Error("create duplicate switch did not yield an error")
	}
}

func TestLogicalPorts(t *testing.T) {
	ovsdbClient := NewFakeOvsdbClient()

	scoreFun := func(left, right interface{}) int {
		ovsdbPort := left.(LPort)
		localPort := right.(LPort)

		switch {
		case ovsdbPort.Name != localPort.Name:
			return -1
		case !reflect.DeepEqual(ovsdbPort.Addresses,
			localPort.Addresses):
			return -1
		default:
			return 0
		}
	}

	checkCorrectness := func(ovsdbLPorts []LPort, localLPorts ...LPort) {
		pair, _, _ := join.Join(ovsdbLPorts, localLPorts, scoreFun)
		if len(pair) != len(localLPorts) {
			t.Error("Local LPort do not match ovsdb LPorts.")
		}
	}

	// Create new switch.
	lswitch := "test-switch"
	if err := ovsdbClient.CreateLogicalSwitch(lswitch); err != nil {
		t.Error(err)
	}

	// Nothing happens yet. It should have zero logical port to be listed.
	ovsdbLPorts, err := ovsdbClient.ListLogicalPorts(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbLPorts) != 0 {
		t.Errorf("expected 0 logical port. Got %d instead.", len(ovsdbLPorts))
	}

	// Create logical port.
	name1, mac1, ip1 := "lp1", "00:00:00:00:00:00", "0.0.0.0"
	lport1 := LPort{
		Name:      "lp1",
		Addresses: []string{fmt.Sprintf("%s %s", mac1, ip1)},
	}
	if err := ovsdbClient.CreateLogicalPort(lswitch, name1, mac1, ip1); err != nil {
		t.Error(err)
	}

	// It should now have one logical port to be listed.
	ovsdbLPorts, err = ovsdbClient.ListLogicalPorts(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbLPorts) != 1 {
		t.Errorf("expected 1 logical port. Got %d instead.", len(ovsdbLPorts))
	}
	ovsdbLPort1 := ovsdbLPorts[0]

	checkCorrectness(ovsdbLPorts, lport1)

	// Create a second logical port.
	name2, mac2, ip2 := "lp2", "00:00:00:00:00:01", "0.0.0.1"
	lport2 := LPort{
		Name:      "lp2",
		Addresses: []string{fmt.Sprintf("%s %s", mac2, ip2)},
	}
	if err := ovsdbClient.CreateLogicalPort(lswitch, name2, mac2, ip2); err != nil {
		t.Error(err)
	}

	// It should now have two logical ports to be listed.
	ovsdbLPorts, err = ovsdbClient.ListLogicalPorts(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbLPorts) != 2 {
		t.Errorf("expected 2 logical port. Got %d instead.", len(ovsdbLPorts))
	}

	checkCorrectness(ovsdbLPorts, lport1, lport2)

	// Delete the first logical port.
	if err := ovsdbClient.DeleteLogicalPort(lswitch, ovsdbLPort1); err != nil {
		t.Error(err)
	}

	// It should now have one logical port to be listed.
	ovsdbLPorts, err = ovsdbClient.ListLogicalPorts(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbLPorts) != 1 {
		t.Errorf("expected 1 logical port. Got %d instead.", len(ovsdbLPorts))
	}
	ovsdbLPort2 := ovsdbLPorts[0]

	checkCorrectness(ovsdbLPorts, lport2)

	// Delete the last one as well.
	if err := ovsdbClient.DeleteLogicalPort(lswitch, ovsdbLPort2); err != nil {
		t.Error(err)
	}

	// It should now have one logical port to be listed.
	ovsdbLPorts, err = ovsdbClient.ListLogicalPorts(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbLPorts) != 0 {
		t.Errorf("expected 0 logical port. Got %d instead.", len(ovsdbLPorts))
	}
}

func TestACLs(t *testing.T) {
	ovsdbClient := NewFakeOvsdbClient()

	key := func(val interface{}) interface{} {
		return val.(Acl).Core
	}

	checkCorrectness := func(ovsdbACLs []Acl, localACLs ...Acl) {
		pair, _, _ := join.HashJoin(AclSlice(ovsdbACLs), AclSlice(localACLs),
			key, key)
		if len(pair) != len(localACLs) {
			t.Error("Local ACLs do not match ovsdbACLs.")
		}
	}

	// Create new switch.
	lswitch := "test-switch"
	if err := ovsdbClient.CreateLogicalSwitch(lswitch); err != nil {
		t.Error(err)
	}

	// Create one ACL rule.
	localCore1 := AclCore{
		Priority:  1,
		Direction: "from-lport",
		Match:     "0.0.0.0",
		Action:    "allow",
	}

	localACL1 := Acl{
		Core: localCore1,
		Log:  false,
	}

	if err := ovsdbClient.CreateACL(lswitch, localCore1.Direction,
		localCore1.Priority, localCore1.Match, localCore1.Action); err != nil {
		t.Error(err)
	}

	// It should now have one ACL entry to be listed.
	ovsdbACLs, err := ovsdbClient.ListACLs(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbACLs) != 1 {
		t.Errorf("expected 1 ACL entry. Got %d instead.", len(ovsdbACLs))
	}
	ovsdbACL1 := ovsdbACLs[0]

	checkCorrectness(ovsdbACLs, localACL1)

	// Create one more ACL rule.
	localCore2 := AclCore{
		Priority:  2,
		Direction: "from-lport",
		Match:     "0.0.0.1",
		Action:    "drop",
	}
	localACL2 := Acl{
		Core: localCore2,
		Log:  false,
	}

	if err := ovsdbClient.CreateACL(lswitch, localCore2.Direction,
		localCore2.Priority, localCore2.Match, localCore2.Action); err != nil {
		t.Error(err)
	}

	// It should now have two ACL entries to be listed.
	ovsdbACLs, err = ovsdbClient.ListACLs(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbACLs) != 2 {
		t.Errorf("expected 2 ACL entry. Got %d instead.", len(ovsdbACLs))
	}

	checkCorrectness(ovsdbACLs, localACL1, localACL2)

	// Delete the first ACL rule.
	if err := ovsdbClient.DeleteACL(lswitch, ovsdbACL1); err != nil {
		t.Error(err)
	}

	// It should now have only one ACL entry to be listed.
	ovsdbACLs, err = ovsdbClient.ListACLs(lswitch)
	if err != nil {
		t.Error(err)
	}

	ovsdbACL2 := ovsdbACLs[0]

	if len(ovsdbACLs) != 1 {
		t.Errorf("expected 1 ACL entry. Got %d instead.", len(ovsdbACLs))
	}

	checkCorrectness(ovsdbACLs, localACL2)

	// Delete the other ACL rule.
	if err := ovsdbClient.DeleteACL(lswitch, ovsdbACL2); err != nil {
		t.Error(err)
	}

	// It should now have only one ACL entry to be listed.
	ovsdbACLs, err = ovsdbClient.ListACLs(lswitch)
	if err != nil {
		t.Error(err)
	}

	if len(ovsdbACLs) != 0 {
		t.Errorf("expected 0 ACL entry. Got %d instead.", len(ovsdbACLs))
	}

}

func TestInterfaces(t *testing.T) {
	ovsdbClient := NewFakeOvsdbClient()

	// Create new switch.
	lswitch1 := "test-switch-1"
	if err := ovsdbClient.CreateLogicalSwitch(lswitch1); err != nil {
		t.Error(err)
	}

	key := func(val interface{}) interface{} {
		iface := val.(Interface)
		return struct {
			name, bridge string
		}{
			name:   iface.Name,
			bridge: iface.Bridge,
		}
	}

	checkCorrectness := func(ovsdbIfaces []Interface, localIfaces ...Interface) {
		pair, _, _ := join.HashJoin(InterfaceSlice(ovsdbIfaces),
			InterfaceSlice(localIfaces), key, key)
		if len(pair) != len(ovsdbIfaces) {
			t.Errorf("local interfaces do not match ovsdb interfaces.")
		}
	}

	// Create a new Bridge. In quilt this is usually done in supervisor.
	_, err := ovsdbClient.transact("Open_vSwitch", ovs.Operation{
		Op:    "insert",
		Table: "Bridge",
		Row:   map[string]interface{}{"name": lswitch1},
	})
	if err != nil {
		t.Error(err)
	}

	// Create one interface.
	iface1 := Interface{
		Name:   "iface1",
		Bridge: lswitch1,
	}

	if err := ovsdbClient.CreateInterface(iface1.Bridge, iface1.Name); err != nil {
		t.Error(err)
	}

	ifaces, err := ovsdbClient.ListInterfaces()
	if err != nil {
		t.Error(err)
	}

	if len(ifaces) != 1 {
		t.Errorf("expected 1 interfaces. Got %d instead.", len(ifaces))
	}

	checkCorrectness(ifaces, iface1)

	// Now create a new switch and bridge. Attach one new interface to them.

	// Create new switch.
	lswitch2 := "test-switch-2"
	if err := ovsdbClient.CreateLogicalSwitch(lswitch2); err != nil {
		t.Error(err)
	}

	// Create a new Bridge.
	_, err = ovsdbClient.transact("Open_vSwitch", ovs.Operation{
		Op:    "insert",
		Table: "Bridge",
		Row:   map[string]interface{}{"name": lswitch2},
	})
	if err != nil {
		t.Error(err)
	}

	// Create a new interface.
	iface2 := Interface{
		Name:   "iface2",
		Bridge: lswitch2,
	}

	if err := ovsdbClient.CreateInterface(iface2.Bridge, iface2.Name); err != nil {
		t.Error(err)
	}

	ifaces, err = ovsdbClient.ListInterfaces()
	if err != nil {
		t.Error(err)
	}

	if len(ifaces) != 2 {
		t.Errorf("expected 2 interfaces. Got %d instead.", len(ifaces))
	}

	checkCorrectness(ifaces, iface1, iface2)

	iface1, iface2 = ifaces[0], ifaces[1]

	// Delete interface 1.
	if err := ovsdbClient.DeleteInterface(iface1); err != nil {
		t.Error(err)
	}

	ifaces, err = ovsdbClient.ListInterfaces()
	if err != nil {
		t.Error(err)
	}

	if len(ifaces) != 1 {
		t.Errorf("expected 1 interfaces. Got %d instead.", len(ifaces))
	}

	checkCorrectness(ifaces, iface2)

	// Delete interface 2.
	if err := ovsdbClient.DeleteInterface(iface2); err != nil {
		t.Error(err)
	}

	ifaces, err = ovsdbClient.ListInterfaces()
	if err != nil {
		t.Error(err)
	}

	if len(ifaces) != 0 {
		t.Errorf("expected 0 interfaces. Got %d instead.", len(ifaces))
	}

	// Test ModifyInterface. We do this by creating an interface with type peer,
	// attach a mac address to it, and add external_ids.
	iface := Interface{
		Name:        "test-modify-iface",
		Peer:        "lolz",
		AttachedMAC: "00:00:00:00:00:00",
		Bridge:      lswitch1,
		Type:        "patch",
	}

	if err := ovsdbClient.CreateInterface(iface.Bridge, iface.Name); err != nil {
		t.Error(err)
	}

	if err := ovsdbClient.ModifyInterface(iface); err != nil {
		t.Error(err)
	}

	ifaces, err = ovsdbClient.ListInterfaces()
	if err != nil {
		t.Error(err)
	}

	ovsdbIface := ifaces[0]
	if ovsdbIface.Name != iface.Name ||
		ovsdbIface.Peer != iface.Peer ||
		ovsdbIface.AttachedMAC != iface.AttachedMAC ||
		ovsdbIface.Bridge != iface.Bridge ||
		ovsdbIface.Type != iface.Type {
		t.Errorf("Modify interface not match. Local: %+v. Remote: %+v.",
			iface, ovsdbIface)
	}

}

func TestBridgeMac(t *testing.T) {
	ovsdbClient := NewFakeOvsdbClient()

	// Create new switch and a new bridge.
	lswitch := "test-switch"
	if err := ovsdbClient.CreateLogicalSwitch(lswitch); err != nil {
		t.Error(err)
	}

	_, err := ovsdbClient.transact("Open_vSwitch", ovs.Operation{
		Op:    "insert",
		Table: "Bridge",
		Row:   map[string]interface{}{"name": lswitch},
	})
	if err != nil {
		t.Error(err)
	}

	mac := "00:00:00:00:00:00"
	if err := ovsdbClient.SetBridgeMac(lswitch, mac); err != nil {
		t.Error(err)
	}

	bridgeReply, err := ovsdbClient.transact("Open_vSwitch", ovs.Operation{
		Op:    "select",
		Table: "Bridge",
		Where: noCondition(),
	})
	if err != nil {
		t.Error(err)
	}

	bridge := bridgeReply[0].Rows[0]
	otherConfig := bridge["other_config"].([]interface{})[1].(map[string]interface{})

	if otherConfig["hwaddr"] != mac {
		t.Error("failed to set bridge mac.")
	}
}

type AclSlice []Acl

func (acls AclSlice) Get(i int) interface{} {
	return acls[i]
}

func (acls AclSlice) Len() int {
	return len(acls)
}

type LPortSlice []LPort

func (lports LPortSlice) Get(i int) interface{} {
	return lports[i]
}

func (lports LPortSlice) Len() int {
	return len(lports)
}
