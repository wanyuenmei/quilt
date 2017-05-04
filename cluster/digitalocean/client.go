package digitalocean

import (
	"context"

	"github.com/digitalocean/godo"
)

// client for DigitalOcean's API. Used for unit testing.
type client interface {
	CreateDroplet(*godo.DropletCreateRequest) (*godo.Droplet, *godo.Response, error)
	DeleteDroplet(int) (*godo.Response, error)
	GetDroplet(int) (*godo.Droplet, *godo.Response, error)
	ListDroplets(*godo.ListOptions) ([]godo.Droplet, *godo.Response, error)

	ListFloatingIPs(*godo.ListOptions) ([]godo.FloatingIP, *godo.Response, error)
	AssignFloatingIP(string, int) (*godo.Action, *godo.Response, error)
	UnassignFloatingIP(string) (*godo.Action, *godo.Response, error)
}

type doClient struct {
	droplets          godo.DropletsService
	floatingIPs       godo.FloatingIPsService
	floatingIPActions godo.FloatingIPActionsService
}

func (client doClient) CreateDroplet(req *godo.DropletCreateRequest) (*godo.Droplet,
	*godo.Response, error) {
	return client.droplets.Create(context.Background(), req)
}

func (client doClient) DeleteDroplet(id int) (*godo.Response, error) {
	return client.droplets.Delete(context.Background(), id)
}

func (client doClient) GetDroplet(id int) (*godo.Droplet, *godo.Response, error) {
	return client.droplets.Get(context.Background(), id)
}

func (client doClient) ListDroplets(opt *godo.ListOptions) ([]godo.Droplet,
	*godo.Response, error) {
	return client.droplets.List(context.Background(), opt)
}

func (client doClient) ListFloatingIPs(opt *godo.ListOptions) ([]godo.FloatingIP,
	*godo.Response, error) {
	return client.floatingIPs.List(context.Background(), opt)
}

func (client doClient) AssignFloatingIP(ip string, id int) (*godo.Action,
	*godo.Response, error) {
	return client.floatingIPActions.Assign(context.Background(), ip, id)
}

func (client doClient) UnassignFloatingIP(ip string) (*godo.Action, *godo.Response,
	error) {
	return client.floatingIPActions.Unassign(context.Background(), ip)
}
