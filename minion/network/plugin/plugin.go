package plugin

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/NetSys/quilt/minion/ipdef"
	"github.com/NetSys/quilt/minion/network"

	dnet "github.com/docker/go-plugins-helpers/network"
)

const (
	// NetworkName is the name of the network driver plugin.
	NetworkName = "quilt"

	pluginDir = "/run/docker/plugins"
	innerVeth = "eth"
)

var (
	networkSocket = NetworkName + ".sock"
	pluginSocket  = filepath.Join(pluginDir, networkSocket)
)

type driver struct{}

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

	mac := req.Interface.MacAddress
	if mac != "" {
		return nil, fmt.Errorf("expected no mac in request, got mac=%s", mac)
	}

	if err := expectNoEndpoint(req.EndpointID); err != nil {
		return nil, err
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
	if err := expectEndpoint(req.EndpointID); err != nil {
		return nil, err
	}
	return &dnet.InfoResponse{}, nil
}

// DeleteEndpoint will do nothing, but checks for the error condition of deleting a
// non-existent endpoint.
func (d driver) DeleteEndpoint(req *dnet.DeleteEndpointRequest) error {
	return expectEndpoint(req.EndpointID)
}

// Join creates a Veth pair for the given endpoint ID, returning the interface info.
func (d driver) Join(req *dnet.JoinRequest) (*dnet.JoinResponse, error) {
	if err := expectNoEndpoint(req.EndpointID); err != nil {
		return nil, err
	}

	// We just need to create the Veth and tell Docker where it should go; Docker
	// will take care of moving it into the container and renaming it.
	tempPeer, err := network.AddVeth(req.EndpointID)
	if err != nil {
		network.DelVeth(req.EndpointID) // Just in case
		return nil, err
	}

	resp := &dnet.JoinResponse{}
	resp.InterfaceName = dnet.InterfaceName{SrcName: tempPeer, DstPrefix: innerVeth}
	resp.Gateway = ipdef.GatewayIP.String()
	return resp, nil
}

// Leave destroys a veth pair for the given endpoint ID.
func (d driver) Leave(req *dnet.LeaveRequest) error {
	if err := expectEndpoint(req.EndpointID); err != nil {
		return err
	}

	return network.DelVeth(req.EndpointID)
}

// CreateNetwork is a noop.
func (d driver) CreateNetwork(req *dnet.CreateNetworkRequest) error {
	return nil
}

// AllocateNetwork is a noop.
func (d driver) AllocateNetwork(req *dnet.AllocateNetworkRequest) (
	*dnet.AllocateNetworkResponse, error) {

	return &dnet.AllocateNetworkResponse{}, nil
}

// DeleteNetwork is a noop.
func (d driver) DeleteNetwork(req *dnet.DeleteNetworkRequest) error {
	return nil
}

// FreeNetwork is a noop.
func (d driver) FreeNetwork(req *dnet.FreeNetworkRequest) error {
	return nil
}

// DiscoverNew is a noop.
func (d driver) DiscoverNew(n *dnet.DiscoveryNotification) error {
	return nil
}

// DiscoverDelete is a noop.
func (d driver) DiscoverDelete(n *dnet.DiscoveryNotification) error {
	return nil
}

// ProgramExternalConnectivity is a noop.
func (d driver) ProgramExternalConnectivity(
	req *dnet.ProgramExternalConnectivityRequest) error {

	return nil
}

// RevokeExternalConnectivity is a noop.
func (d driver) RevokeExternalConnectivity(
	req *dnet.RevokeExternalConnectivityRequest) error {

	return nil
}

func expectEndpoint(eid string) error {
	if ok, err := endpointExists(eid); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("endpoint %s doesn't exists", eid)
	}

	return nil
}

func expectNoEndpoint(eid string) error {
	if ok, err := endpointExists(eid); err != nil {
		return err
	} else if ok {
		return fmt.Errorf("endpoint %s already exists", eid)
	}

	return nil
}

func endpointExists(eid string) (bool, error) {
	_, name := network.VethPairNames(eid)
	return network.LinkExists("", name)
}

var rand32 = rand.Uint32
