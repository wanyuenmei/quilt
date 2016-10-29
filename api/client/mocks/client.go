package mocks

import (
	"github.com/NetSys/quilt/db"
)

// Client implements a mocked version of a Quilt client.
type Client struct {
	MachineReturn   []db.Machine
	ContainerReturn []db.Container
	EtcdReturn      []db.Etcd
	HostReturn      string
	DeployArg       string
}

// QueryMachines retrieves the machines tracked by the Quilt daemon.
func (c *Client) QueryMachines() ([]db.Machine, error) {
	return c.MachineReturn, nil
}

// QueryContainers retrieves the containers tracked by the Quilt daemon.
func (c *Client) QueryContainers() ([]db.Container, error) {
	return c.ContainerReturn, nil
}

// QueryEtcd retrieves the etcd information tracked by the Quilt daemon.
func (c *Client) QueryEtcd() ([]db.Etcd, error) {
	return c.EtcdReturn, nil
}

// Close the grpc connection.
func (c *Client) Close() error {
	return nil
}

// Deploy makes a request to the Quilt daemon to deploy the given deployment.
func (c *Client) Deploy(depl string) error {
	c.DeployArg = depl
	return nil
}

// Host returns the server address the Client is connected to.
func (c *Client) Host() string {
	return c.HostReturn
}
