package google

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	compute "google.golang.org/api/compute/v1"

	"github.com/quilt/quilt/util"
)

//go:generate mockery -inpkg -testonly -name=client
type client interface {
	GetInstance(zone, id string) (*compute.Instance, error)
	ListInstances(zone string, opts apiOptions) (*compute.InstanceList,
		error)
	InsertInstance(zone string, instance *compute.Instance) (
		*compute.Operation, error)
	DeleteInstance(zone, operation string) (*compute.Operation, error)
	AddAccessConfig(zone, instance, networkInterface string,
		accessConfig *compute.AccessConfig) (*compute.Operation, error)
	DeleteAccessConfig(zone, instance, accessConfig,
		networkInterface string) (*compute.Operation, error)
	GetZoneOperation(zone, operation string) (*compute.Operation, error)
	GetGlobalOperation(operation string) (*compute.Operation, error)
	ListFirewalls() (*compute.FirewallList, error)
	InsertFirewall(firewall *compute.Firewall) (
		*compute.Operation, error)
	PatchFirewall(name string, firewall *compute.Firewall) (
		*compute.Operation, error)
	DeleteFirewall(firewall string) (*compute.Operation, error)
	ListNetworks() (*compute.NetworkList, error)
	InsertNetwork(network *compute.Network) (
		*compute.Operation, error)
}

type clientImpl struct {
	gce    *compute.Service
	projID string
}

func newClient() (*clientImpl, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".gce", "quilt.json")
	configStr, err := util.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	service, err := newComputeService(configStr)
	if err != nil {
		return nil, err
	}

	projID, err := getProjectID(configStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get project ID: %s", err)
	}

	return &clientImpl{gce: service, projID: projID}, nil
}

const projectIDKey = "project_id"

func getProjectID(configStr string) (string, error) {
	configFields := map[string]string{}
	if err := json.Unmarshal([]byte(configStr), &configFields); err != nil {
		return "", err
	}

	projID, ok := configFields[projectIDKey]
	if !ok {
		return "", fmt.Errorf("missing field: %s", projectIDKey)
	}

	return projID, nil
}

/**
 * Service: Instances
 */

func (c *clientImpl) GetInstance(zone, id string) (*compute.Instance, error) {
	return c.gce.Instances.Get(c.projID, zone, id).Do()
}

type apiOptions struct {
	filter string
}

func (c *clientImpl) ListInstances(zone string, opts apiOptions) (
	*compute.InstanceList, error) {
	call := c.gce.Instances.List(c.projID, zone)
	if opts.filter != "" {
		call = call.Filter(opts.filter)
	}

	return call.Do()
}

func (c *clientImpl) InsertInstance(zone string, instance *compute.Instance) (
	*compute.Operation, error) {
	return c.gce.Instances.Insert(c.projID, zone, instance).Do()
}

func (c *clientImpl) DeleteInstance(zone, instance string) (*compute.Operation,
	error) {
	return c.gce.Instances.Delete(c.projID, zone, instance).Do()
}

func (c *clientImpl) AddAccessConfig(zone, instance, networkInterface string,
	accessConfig *compute.AccessConfig) (*compute.Operation, error) {
	return c.gce.Instances.AddAccessConfig(c.projID, zone, instance, networkInterface,
		accessConfig).Do()
}

func (c *clientImpl) DeleteAccessConfig(zone, instance, accessConfig,
	networkInterface string) (*compute.Operation, error) {
	return c.gce.Instances.DeleteAccessConfig(c.projID, zone, instance,
		accessConfig, networkInterface).Do()
}

/**
 * Service: ZoneOperations
 */

func (c *clientImpl) GetZoneOperation(zone, operation string) (
	*compute.Operation, error) {
	return c.gce.ZoneOperations.Get(c.projID, zone, operation).Do()
}

/**
 * Service: GlobalOperations
 */

func (c *clientImpl) GetGlobalOperation(operation string) (*compute.Operation,
	error) {
	return c.gce.GlobalOperations.Get(c.projID, operation).Do()
}

/**
 * Service: Firewall
 */

func (c *clientImpl) ListFirewalls() (*compute.FirewallList, error) {
	return c.gce.Firewalls.List(c.projID).Do()
}

func (c *clientImpl) InsertFirewall(firewall *compute.Firewall) (
	*compute.Operation, error) {
	return c.gce.Firewalls.Insert(c.projID, firewall).Do()
}

func (c *clientImpl) PatchFirewall(name string, firewall *compute.Firewall) (
	*compute.Operation, error) {
	return c.gce.Firewalls.Patch(c.projID, name, firewall).Do()
}

func (c *clientImpl) DeleteFirewall(firewall string) (
	*compute.Operation, error) {
	return c.gce.Firewalls.Delete(c.projID, firewall).Do()
}

/**
 * Service: Networks
 */

func (c *clientImpl) ListNetworks() (*compute.NetworkList, error) {
	return c.gce.Networks.List(c.projID).Do()
}

func (c *clientImpl) InsertNetwork(network *compute.Network) (
	*compute.Operation, error) {
	return c.gce.Networks.Insert(c.projID, network).Do()
}
