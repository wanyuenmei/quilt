package plugin

import (
	"fmt"
	"testing"

	"github.com/NetSys/quilt/minion/ipdef"
	"github.com/NetSys/quilt/minion/network"
	"github.com/stretchr/testify/assert"

	dnet "github.com/docker/go-plugins-helpers/network"
)

var (
	m          mock
	delVeth    = network.DelVeth
	addVeth    = network.AddVeth
	linkExists = network.LinkExists

	zero = "0000000000000"
	one  = "1111111111111"
)

type mock map[string]struct{}

func (m mock) fakeDelVeth(s string) error {
	name := ipdef.IFName(s)
	if _, ok := m[name]; !ok {
		return fmt.Errorf("no such veth: %s", name)
	}
	delete(m, name)
	return nil
}

func (m mock) fakeAddVeth(s string) (string, error) {
	name := ipdef.IFName(s)
	if _, ok := m[name]; ok {
		return "", fmt.Errorf("veth exists: %s", name)
	}
	m[name] = struct{}{}
	return "tmpVeth", nil
}

func (m mock) fakeLinkExists(s, t string) (bool, error) {
	_, ok := m[t]
	return ok, nil
}

func setup() {
	m = mock(map[string]struct{}{})
	network.DelVeth = m.fakeDelVeth
	network.AddVeth = m.fakeAddVeth
	network.LinkExists = m.fakeLinkExists
}

func teardown() {
	network.DelVeth = delVeth
	network.AddVeth = addVeth
	network.LinkExists = linkExists
}

func TestGetCapabilities(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	resp, err := d.GetCapabilities()
	assert.NoError(t, err)

	exp := dnet.CapabilitiesResponse{Scope: dnet.LocalScope}
	assert.Equal(t, exp, *resp)
}

func TestCreateEndpoint(t *testing.T) {
	setup()
	defer teardown()

	req := &dnet.CreateEndpointRequest{}
	req.EndpointID = zero
	req.Interface = &dnet.EndpointInterface{
		MacAddress: "00:00:00:00:00:00",
	}

	d := driver{}
	_, err := d.CreateEndpoint(req)
	assert.EqualError(t, err, "invalid IP: ")

	req.Interface.Address = "10.1.0.1"
	_, err = d.CreateEndpoint(req)
	assert.EqualError(t, err, "invalid IP: 10.1.0.1")

	req.Interface.Address = "10.1.0.1/8"
	req.Interface.MacAddress = ""
	expResp := dnet.EndpointInterface{
		MacAddress: ipdef.IPStrToMac("10.1.0.1"),
	}
	resp, err := d.CreateEndpoint(req)
	assert.NoError(t, err)
	assert.Equal(t, expResp, *resp.Interface)

	req.EndpointID = one
	req.Interface.Address = "10.1.0.2/8"
	expResp.MacAddress = ipdef.IPStrToMac("10.1.0.2")
	resp, err = d.CreateEndpoint(req)
	assert.NoError(t, err)
	assert.Equal(t, expResp, *resp.Interface)
}

func TestDeleteEndpoint(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	req := &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"}
	_, err := d.Join(req)
	assert.NoError(t, err)

	delReq := &dnet.DeleteEndpointRequest{EndpointID: zero}
	err = d.DeleteEndpoint(delReq)
	assert.NoError(t, err)

	d.Leave(&dnet.LeaveRequest{EndpointID: zero})
	delReq = &dnet.DeleteEndpointRequest{EndpointID: zero}
	err = d.DeleteEndpoint(delReq)
	assert.EqualError(t, err, "endpoint 0000000000000 doesn't exists")
}

func TestEndpointOperInfo(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	req := &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"}
	_, err := d.Join(req)
	assert.NoError(t, err)

	_, err = d.EndpointInfo(&dnet.InfoRequest{EndpointID: zero})
	assert.NoError(t, err)

	d.Leave(&dnet.LeaveRequest{EndpointID: zero})
	_, err = d.EndpointInfo(&dnet.InfoRequest{EndpointID: one})
	assert.EqualError(t, err, "endpoint 1111111111111 doesn't exists")
}

func TestJoin(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	jreq := &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"}
	resp, err := d.Join(jreq)
	assert.NoError(t, err)

	ifaceName := resp.InterfaceName
	expIFace := dnet.InterfaceName{SrcName: "tmpVeth", DstPrefix: ifacePrefix}
	assert.Equal(t, expIFace, ifaceName)

	jreq = &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker2"}
	_, err = d.Join(jreq)
	assert.EqualError(t, err, "endpoint 0000000000000 already exists")
}

func TestLeave(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	_, err := d.Join(&dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"})
	assert.NoError(t, err)

	err = d.Leave(&dnet.LeaveRequest{EndpointID: zero})
	assert.NoError(t, err)

	err = d.DeleteEndpoint(&dnet.DeleteEndpointRequest{EndpointID: zero})
	assert.EqualError(t, err, "endpoint 0000000000000 doesn't exists")
}
