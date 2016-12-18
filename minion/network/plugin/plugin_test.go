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
	assert.Nil(t, err)

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
	assert.NotNil(t, err)

	req.Interface.Address = "10.1.0.1"
	_, err = d.CreateEndpoint(req)
	assert.NotNil(t, err)

	req.Interface.Address = "10.1.0.1/8"
	req.Interface.MacAddress = ""
	expResp := dnet.EndpointInterface{
		MacAddress: ipdef.IPStrToMac("10.1.0.1"),
	}
	resp, err := d.CreateEndpoint(req)
	assert.Nil(t, err)
	assert.Equal(t, expResp, *resp.Interface)

	req.EndpointID = one
	req.Interface.Address = "10.1.0.2/8"
	expResp.MacAddress = ipdef.IPStrToMac("10.1.0.2")
	resp, err = d.CreateEndpoint(req)
	assert.Nil(t, err)
	assert.Equal(t, expResp, *resp.Interface)
}

func TestDeleteEndpoint(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	req := &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"}
	d.Join(req)

	delReq := &dnet.DeleteEndpointRequest{EndpointID: zero}
	err := d.DeleteEndpoint(delReq)
	assert.Nil(t, err)

	d.Leave(&dnet.LeaveRequest{EndpointID: zero})
	delReq = &dnet.DeleteEndpointRequest{EndpointID: zero}
	err = d.DeleteEndpoint(delReq)
	assert.NotNil(t, err)
}

func TestEndpointOperInfo(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	req := &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"}
	d.Join(req)
	_, err := d.EndpointInfo(&dnet.InfoRequest{EndpointID: zero})
	assert.Nil(t, err)

	d.Leave(&dnet.LeaveRequest{EndpointID: zero})
	_, err = d.EndpointInfo(&dnet.InfoRequest{EndpointID: one})
	assert.NotNil(t, err)
}

func TestJoin(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	jreq := &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"}
	resp, err := d.Join(jreq)
	assert.Nil(t, err)

	ifaceName := resp.InterfaceName
	expIFace := dnet.InterfaceName{SrcName: "tmpVeth", DstPrefix: innerVeth}
	assert.Equal(t, expIFace, ifaceName)

	jreq = &dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker2"}
	_, err = d.Join(jreq)
	assert.NotNil(t, err)
}

func TestLeave(t *testing.T) {
	setup()
	defer teardown()

	d := driver{}
	d.Join(&dnet.JoinRequest{EndpointID: zero, SandboxKey: "/test/docker0"})

	err := d.Leave(&dnet.LeaveRequest{EndpointID: zero})
	assert.Nil(t, err)

	err = d.DeleteEndpoint(&dnet.DeleteEndpointRequest{EndpointID: zero})
	assert.NotNil(t, err)
}

func TestNoop(t *testing.T) {
	setup()
	defer teardown()
	d := driver{}

	err := d.CreateNetwork(&dnet.CreateNetworkRequest{})
	assert.Nil(t, err)

	err = d.FreeNetwork(&dnet.FreeNetworkRequest{})
	assert.Nil(t, err)

	err = d.DiscoverNew(&dnet.DiscoveryNotification{})
	assert.Nil(t, err)

	err = d.DiscoverDelete(&dnet.DiscoveryNotification{})
	assert.Nil(t, err)

	err = d.ProgramExternalConnectivity(&dnet.ProgramExternalConnectivityRequest{})
	assert.Nil(t, err)

	err = d.RevokeExternalConnectivity(&dnet.RevokeExternalConnectivityRequest{})
	assert.Nil(t, err)

	resp, err := d.AllocateNetwork(&dnet.AllocateNetworkRequest{})
	assert.Nil(t, err)

	if resp.Options != nil && len(resp.Options) > 0 {
		t.Fatalf("AllocateNetwork responded with non-empty response: %v", *resp)
	}

	err = d.DeleteNetwork(&dnet.DeleteNetworkRequest{})
	assert.Nil(t, err)
}
