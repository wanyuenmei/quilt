package plugin

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/NetSys/quilt/minion/ipdef"

	dnet "github.com/docker/go-plugins-helpers/network"
	"github.com/vishvananda/netlink"
)

const (
	// NetworkName is the name of the network driver plugin.
	NetworkName = "quilt"

	pluginDir   = "/run/docker/plugins"
	ifacePrefix = "eth"
)

var (
	networkSocket = NetworkName + ".sock"
	pluginSocket  = filepath.Join(pluginDir, networkSocket)
)

type driver struct{}

const mtu int = 1400

// Run runs the network driver and starts the server to listen for requests. It will
// block until the server socket has been created.
func Run() {
	h := dnet.NewHandler(driver{})

	go func() {
		err := h.ServeUnix("root", pluginSocket)
		if err != nil {
			// If the driver fails to start, we can't boot any containers,
			// so we may as well panic.
			panic("Failed to serve driver socket server")
		}
	}()

	// The ServeUnix function that handles the plugin socket won't return until
	// the socket is closed, so we can't know exactly when the socket has been
	// created. In order to prevent a race condition where Docker attempts to use
	// the plugin before the socket is up, we simply wait until the socket file
	// exists.
	for {
		if _, err := os.Stat(pluginSocket); err == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// GetCapabilities returns the capabilities of this network driver.
func (d driver) GetCapabilities() (*dnet.CapabilitiesResponse, error) {
	return &dnet.CapabilitiesResponse{Scope: dnet.LocalScope}, nil
}

// CreateEndpoint acknowledges the request, but does not actually do anything.
func (d driver) CreateEndpoint(req *dnet.CreateEndpointRequest) (
	*dnet.CreateEndpointResponse, error) {

	addr, _, err := net.ParseCIDR(req.Interface.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid IP: %s", req.Interface.Address)
	}

	if _, err := getOuterLink(req.EndpointID); err == nil {
		return nil, fmt.Errorf("endpoint %s exists", req.EndpointID)
	}

	resp := &dnet.CreateEndpointResponse{
		Interface: &dnet.EndpointInterface{
			MacAddress: ipdef.IPToMac(addr),
		},
	}
	return resp, nil
}

// EndpointInfo will return an error if the endpoint does not exist.
func (d driver) EndpointInfo(req *dnet.InfoRequest) (*dnet.InfoResponse, error) {
	if _, err := getOuterLink(req.EndpointID); err != nil {
		return nil, err
	}
	return &dnet.InfoResponse{}, nil
}

// DeleteEndpoint will do nothing, but checks for the error condition of deleting a
// non-existent endpoint.
func (d driver) DeleteEndpoint(req *dnet.DeleteEndpointRequest) error {
	_, err := getOuterLink(req.EndpointID)
	return err
}

// Join creates a Veth pair for the given endpoint ID, returning the interface info.
func (d driver) Join(req *dnet.JoinRequest) (*dnet.JoinResponse, error) {
	// We just need to create the Veth and tell Docker where it should go; Docker
	// will take care of moving it into the container and renaming it.
	outer := ipdef.IFName(req.EndpointID)
	inner := ipdef.IFName("tmp_" + req.EndpointID)
	err := linkAdd(&netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: outer, MTU: mtu},
		PeerName:  inner,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create veth: %s", err)
	}

	outerLink, err := getOuterLink(req.EndpointID)
	if err != nil {
		return nil, fmt.Errorf("failed to find interface %s: %s", outer, err)
	}

	err = linkSetUp(outerLink)
	if err != nil {
		return nil, fmt.Errorf("failed to bring up link %s: %s", outer, err)
	}

	resp := &dnet.JoinResponse{}
	resp.Gateway = ipdef.GatewayIP.String()
	resp.InterfaceName = dnet.InterfaceName{SrcName: inner, DstPrefix: ifacePrefix}
	return resp, nil
}

// Leave destroys a veth pair for the given endpoint ID.
func (d driver) Leave(req *dnet.LeaveRequest) error {
	outer, err := getOuterLink(req.EndpointID)
	if err != nil {
		return fmt.Errorf("failed to find link %s: %s", outer, err)
	}

	if err := linkDel(outer); err != nil {
		return fmt.Errorf("failed to delete link %s: %s", outer, err)
	}
	return nil
}

func getOuterLink(eid string) (netlink.Link, error) {
	return linkByName(ipdef.IFName(eid))
}

// Mock variables for unit testing
var linkAdd = netlink.LinkAdd
var linkDel = netlink.LinkDel
var linkSetUp = netlink.LinkSetUp
var linkByName = netlink.LinkByName
