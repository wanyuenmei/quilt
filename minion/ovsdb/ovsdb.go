//go:generate mockery -name=transact -inpkg
//go:generate mockery -name=Client

package ovsdb

import (
	"errors"
	"fmt"
	"math"

	ovs "github.com/socketplane/libovsdb"
)

type transact interface {
	Transact(db string, operation ...ovs.Operation) ([]ovs.OperationResult, error)
	Disconnect()
}

// Client is a connection to the ovsdb-server database.
type Client interface {
	CreateLogicalSwitch(lswitch string) error
	LogicalSwitchExists(lswitch string) (bool, error)
	ListSwitchPorts() ([]SwitchPort, error)
	CreateSwitchPort(lswitch string, lport SwitchPort) error
	DeleteSwitchPort(lswitch string, lport SwitchPort) error

	ListACLs() ([]ACL, error)
	CreateACL(lswitch, direction string, priority int, match, action string) error
	DeleteACL(lswitch string, ovsdbACL ACL) error

	ListAddressSets() ([]AddressSet, error)
	CreateAddressSet(name string, addresses []string) error
	DeleteAddressSet(name string) error

	OpenFlowPorts() (map[string]int, error)

	Disconnect()
}

type client struct {
	transact
}

// SwitchPort is a logical switch port in OVN.
type SwitchPort struct {
	uuid      ovs.UUID
	Name      string
	Type      string
	Addresses []string
	Options   map[string]string
}

// SwitchPortSlice is a wrapper around []SwitchPort so it can be used in joins
type SwitchPortSlice []SwitchPort

// Interface is a linux device attached to OVS.
type Interface struct {
	Name        string
	Peer        string
	AttachedMAC string
	IfaceID     string
	Bridge      string
	Type        string
}

// ACL is a firewall rule in OVN.
type ACL struct {
	uuid ovs.UUID
	Core ACLCore
	Log  bool
}

// ACLCore is the actual ACL rule that will be matched, without various OVSDB metadata
// found in ACL.
type ACLCore struct {
	Priority  int
	Direction string
	Match     string
	Action    string
}

type row map[string]interface{}
type mutation interface{}

// Base on RFC 7047, empty condition should return all rows from a table. However,
// libovsdb does not seem to support that yet. This is the simplist, most common solution
// to it.
var noCondition = []interface{}{ovs.NewCondition("_uuid", "!=", ovs.UUID{GoUUID: "_"})}

func newCondition(column, function string, value interface{}) []interface{} {
	return []interface{}{ovs.NewCondition(column, function, value)}
}

// Open creates a new Ovsdb connection.
// It's stored in a variable so we can mock it out for the unit tests.
var Open = func() (Client, error) {
	odb, err := ovs.Connect("127.0.0.1", 6640)
	return client{odb}, err
}

// CreateLogicalSwitch creates a new logical switch in OVN.
func (ovsdb client) CreateLogicalSwitch(lswitch string) error {
	insertOp := ovs.Operation{
		Op:    "insert",
		Table: "Logical_Switch",
		Row:   map[string]interface{}{"name": lswitch},
	}

	results, err := ovsdb.Transact("OVN_Northbound", insertOp)
	if err != nil {
		return fmt.Errorf("transaction error: creating switch %s: %s",
			lswitch, err)
	}
	return errorCheck(results, 1)
}

func (ovsdb client) LogicalSwitchExists(lswitch string) (bool, error) {
	matches, err := ovsdb.Transact("OVN_Northbound", ovs.Operation{
		Op:    "select",
		Table: "Logical_Switch",
		Where: newCondition("name", "==", lswitch),
	})
	if err != nil {
		return false, fmt.Errorf(
			"transaction error: listing logical switch: %s", err)
	}
	return len(matches) > 0 && len(matches[0].Rows) > 0, nil
}

// ListSwitchPorts lists the logical ports in OVN.
func (ovsdb client) ListSwitchPorts() ([]SwitchPort, error) {
	portReply, err := ovsdb.Transact("OVN_Northbound", ovs.Operation{
		Op:    "select",
		Table: "Logical_Switch_Port",
		Where: noCondition,
	})
	if err != nil {
		return nil, fmt.Errorf("transaction error: listing switch ports: %s", err)
	}

	var result []SwitchPort
	for _, row := range portReply[0].Rows {
		port, err := parseLogicalSwitchPort(row)
		if err != nil {
			return nil, fmt.Errorf("malformed switch port: %s", err)
		}
		result = append(result, port)
	}
	return result, nil
}

func parseLogicalSwitchPort(port row) (SwitchPort, error) {
	options, err := ovsStringMapToMap(port["options"])
	if err != nil {
		return SwitchPort{}, fmt.Errorf("malformed options: %s", err)
	}
	return SwitchPort{
		uuid:      ovsUUIDFromRow(port),
		Name:      port["name"].(string),
		Type:      port["type"].(string),
		Addresses: ovsStringSetToSlice(port["addresses"]),
		Options:   options,
	}, nil
}

// CreateSwitchPort creates a new logical port in OVN.
func (ovsdb client) CreateSwitchPort(lswitch string, lport SwitchPort) error {
	portRow := map[string]interface{}{
		"name": lport.Name,
		"type": lport.Type,
	}

	if len(lport.Addresses) != 0 {
		portRow["addresses"] = newOvsSet(lport.Addresses)
	}

	if lport.Options != nil {
		portRow["options"] = newOvsMap(lport.Options)
	}

	insertOp := ovs.Operation{
		Op:       "insert",
		Table:    "Logical_Switch_Port",
		Row:      portRow,
		UUIDName: "qlsportadd",
	}

	mutateOp := ovs.Operation{
		Op:    "mutate",
		Table: "Logical_Switch",
		Mutations: []interface{}{
			newMutation("ports", "insert", ovs.UUID{GoUUID: "qlsportadd"}),
		},
		Where: newCondition("name", "==", lswitch),
	}

	results, err := ovsdb.Transact("OVN_Northbound", insertOp, mutateOp)
	if err != nil {
		return fmt.Errorf("transaction error: creating switch port %s on %s: %s",
			lport.Name, lswitch, err)
	}

	return errorCheck(results, 2)
}

// DeleteSwitchPort removes a logical port from OVN.
func (ovsdb client) DeleteSwitchPort(lswitch string, lport SwitchPort) error {
	deleteOp := ovs.Operation{
		Op:    "delete",
		Table: "Logical_Switch_Port",
		Where: newCondition("_uuid", "==", lport.uuid),
	}

	mutateOp := ovs.Operation{
		Op:        "mutate",
		Table:     "Logical_Switch",
		Mutations: []interface{}{newMutation("ports", "delete", lport.uuid)},
		Where:     newCondition("name", "==", lswitch),
	}

	results, err := ovsdb.Transact("OVN_Northbound", deleteOp, mutateOp)
	if err != nil {
		return fmt.Errorf("transaction error: deleting switch port %s on %s: %s",
			lport.Name, lswitch, err)
	}
	return errorCheck(results, 2)
}

// ListACLs lists the access control rules in OVN.
func (ovsdb client) ListACLs() ([]ACL, error) {
	aclReply, err := ovsdb.Transact("OVN_Northbound", ovs.Operation{
		Op:    "select",
		Table: "ACL",
		Where: noCondition,
	})
	if err != nil {
		return nil, fmt.Errorf("transaction error: listing ACLs: %s", err)
	}

	var result []ACL
	for _, row := range aclReply[0].Rows {
		result = append(result, ACL{
			uuid: ovsUUIDFromRow(row),
			Core: ACLCore{
				Priority:  int(row["priority"].(float64)),
				Direction: row["direction"].(string),
				Match:     row["match"].(string),
				Action:    row["action"].(string),
			},
			Log: row["log"].(bool)})
	}
	return result, nil
}

// CreateACL creates an access control rule in OVN.
//
// direction, unless wildcarded, must be either "from-lport" or "to-lport"
//
// priority, unless wildcarded, must be in [1,32767]
//
// match, unless wildcarded or empty, must be a valid OpenFlow expression
//
// action must be one of {"allow", "allow-related", "drop", "reject"}
//
// direction and match may be wildcarded by passing the value "*". priority may also
// be wildcarded by passing a value less than 0.
func (ovsdb client) CreateACL(lswitch, direction string, priority int,
	match, action string) error {
	aclRow := map[string]interface{}{
		"priority": int(math.Max(0.0, float64(priority))),
		"action":   action,
		"log":      false,
	}
	if direction != "*" {
		aclRow["direction"] = direction
	}
	if match != "*" {
		aclRow["match"] = match
	}

	insertOp := ovs.Operation{
		Op:       "insert",
		Table:    "ACL",
		Row:      aclRow,
		UUIDName: "qacladd",
	}

	mutateOp := ovs.Operation{
		Op:    "mutate",
		Table: "Logical_Switch",
		Mutations: []interface{}{
			newMutation("acls", "insert", ovs.UUID{GoUUID: "qacladd"}),
		},
		Where: newCondition("name", "==", lswitch),
	}

	results, err := ovsdb.Transact("OVN_Northbound", insertOp, mutateOp)
	if err != nil {
		return fmt.Errorf("transaction error: creating ACL on %s: %s",
			lswitch, err)
	}
	return errorCheck(results, 2)
}

// DeleteACL removes an access control rule from OVN.
func (ovsdb client) DeleteACL(lswitch string, ovsdbACL ACL) error {
	deleteOp := ovs.Operation{
		Op:    "delete",
		Table: "ACL",
		Where: newCondition("_uuid", "==", ovsdbACL.uuid),
	}

	mutateOp := ovs.Operation{
		Op:    "mutate",
		Table: "Logical_Switch",
		Mutations: []interface{}{
			newMutation("acls", "delete", ovsdbACL.uuid),
		},
		Where: newCondition("name", "==", lswitch),
	}

	results, err := ovsdb.Transact("OVN_Northbound", deleteOp, mutateOp)
	if err != nil {
		return fmt.Errorf("transaction error: deleting ACL on %s: %s",
			lswitch, err)
	}
	return errorCheck(results, 2)
}

// AddressSet is a named group of IPs in OVN.
type AddressSet struct {
	Name      string
	Addresses []string
}

// ListAddressSets lists the address sets in OVN.
func (ovsdb client) ListAddressSets() ([]AddressSet, error) {
	result := []AddressSet{}

	addressReply, err := ovsdb.Transact("OVN_Northbound", ovs.Operation{
		Op:    "select",
		Table: "Address_Set",
		Where: noCondition,
	})
	if err != nil {
		return nil, fmt.Errorf("transaction error: list address sets: %s", err)
	}

	for _, addr := range addressReply[0].Rows {
		result = append(result, AddressSet{
			Name:      addr["name"].(string),
			Addresses: ovsStringSetToSlice(addr["addresses"]),
		})
	}
	return result, nil
}

// CreateAddressSet creates an address set in OVN.
func (ovsdb client) CreateAddressSet(name string, addresses []string) error {
	addrs := newOvsSet(addresses)
	addressRow := map[string]interface{}{
		"name":      name,
		"addresses": addrs,
	}

	insertOp := ovs.Operation{
		Op:    "insert",
		Table: "Address_Set",
		Row:   addressRow,
	}

	results, err := ovsdb.Transact("OVN_Northbound", insertOp)
	if err != nil {
		return fmt.Errorf("transaction error: creating address set: %s", err)
	}
	return errorCheck(results, 1)
}

// DeleteAddressSet removes an address set from OVN.
func (ovsdb client) DeleteAddressSet(name string) error {
	deleteOp := ovs.Operation{
		Op:    "delete",
		Table: "Address_Set",
		Where: newCondition("name", "==", name),
	}

	results, err := ovsdb.Transact("OVN_Northbound", deleteOp)
	if err != nil {
		return fmt.Errorf("transaction error: deleting address set: %s", err)
	}
	return errorCheck(results, 1)
}

// OpenFlowPorts returns a map from interface name to OpenFlow port number for every
// interface in ovsdb.  Those interfaces without a port number are silently omitted.
func (ovsdb client) OpenFlowPorts() (map[string]int, error) {
	reply, err := ovsdb.Transact("Open_vSwitch", ovs.Operation{
		Op:    "select",
		Table: "Interface",
		Where: noCondition,
	})
	if err != nil {
		return nil, fmt.Errorf("select interface error: %s", err)
	}

	ifaceMap := map[string]int{}
	for _, iface := range reply[0].Rows {
		name, ok := iface["name"].(string)
		if !ok {
			continue
		}

		ofport, ok := iface["ofport"].(float64)
		if !ok {
			continue
		}

		if ofport > 0 {
			ifaceMap[name] = int(ofport)
		}
	}

	return ifaceMap, nil
}

// This does not cover all cases, they should just be added as needed
func newMutation(column, mutator string, value interface{}) mutation {
	switch typedValue := value.(type) {
	case ovs.UUID:
		uuidSlice := []ovs.UUID{typedValue}
		mutateValue := newOvsSet(uuidSlice)
		return ovs.NewMutation(column, mutator, mutateValue)
	default:
		panic(fmt.Sprintf("unhandled value in mutation: value %s, type %s",
			value, typedValue))
	}
}

func ovsStringSetToSlice(oSet interface{}) []string {
	var ret []string
	if t, ok := oSet.([]interface{}); ok && t[0] == "set" {
		for _, v := range t[1].([]interface{}) {
			ret = append(ret, v.(string))
		}
	} else {
		ret = append(ret, oSet.(string))
	}
	return ret
}

func ovsStringMapToMap(oMap interface{}) (map[string]string, error) {
	var ret = make(map[string]string)
	wrap, ok := oMap.([]interface{})
	if !ok || len(wrap) == 0 || wrap[0] != "map" {
		return nil, errors.New("ovs map outermost layer invalid")
	}

	brokenMap, ok := wrap[1].([]interface{})
	if !ok {
		return nil, errors.New("ovs map content invalid")
	}
	for _, kvPair := range brokenMap {
		kvSlice, ok := kvPair.([]interface{})
		if !ok {
			return nil, errors.New("ovs map block must be a slice")
		}
		key, ok := kvSlice[0].(string)
		if !ok {
			return nil, errors.New("ovs map key must be string")
		}
		val, ok := kvSlice[1].(string)
		if !ok {
			return nil, errors.New("ovs map value must be string")
		}
		ret[key] = val
	}
	return ret, nil
}

func ovsUUIDFromRow(row row) ovs.UUID {
	uuid := ovs.UUID{}
	block, ok := row["_uuid"].([]interface{})
	if !ok && len(block) != 2 {
		// This should never happen.
		panic("row does not have valid UUID")
	}
	uuid.GoUUID = block[1].(string)
	return uuid
}

func errorCheck(results []ovs.OperationResult, expectedResponses int) error {
	if len(results) < expectedResponses {
		return errors.New("mismatched responses and operations")
	}
	for i, result := range results {
		if result.Error != "" {
			return fmt.Errorf("operation %d failed due to error: %s: %s",
				i, result.Error, result.Details)
		}
	}
	return nil
}

// Get gets the element at the ith index
func (lps SwitchPortSlice) Get(i int) interface{} {
	return lps[i]
}

// Len returns the length of the slice
func (lps SwitchPortSlice) Len() int {
	return len(lps)
}

func newOvsSet(slice interface{}) *ovs.OvsSet {
	result, err := ovs.NewOvsSet(slice)
	if err != nil {
		panic(err)
	}
	return result
}

func newOvsMap(mp interface{}) *ovs.OvsMap {
	result, err := ovs.NewOvsMap(mp)
	if err != nil {
		panic(err)
	}
	return result
}
