# DigitalOcean

To deploy a DigitalOcean droplet in the `sfo1` zone of size
`512mb` as a `Worker`:

```
deployment.deploy(new Machine({
  provider: "DigitalOcean",
  region: "sfo1",
  size: "512mb",
  role: "Worker"
}));
```

## Setup

1. Create a new key [here](https://cloud.digitalocean.com/settings/api/tokens).
Both read and write permissions are required.

2. Save the key in `~/.digitalocean/key` on the machine that will be running the
Quilt daemon.

## Floating IPs
To assign a floating IP to a machine, simply specify the IP as an attribute. For example,

```
deployment.deploy(new Machine({
  provider: "DigitalOcean",
  region: "sfo1",
  size: "512mb",
  floatingIp: "8.8.8.8",
  role: "Worker"
}));
```

Creating a floating IP is slightly unintuitive. Unless there are already
droplets running, the floating IP tab under "Networking" doesn't allow users to
create floating IPs. However, [this
link](https://cloud.digitalocean.com/networking/floating_ips/datacenter) can be
used to reserve IPs for a specific datacenter. If that link breaks, floating
IPs can always be created by creating a droplet, _then_ assigning it a new
floating IP. The floating IP will still be reserved for use after
disassociating it.

Note that DigitalOcean charges a fee of $.0006/hr for floating IPs that have
been reserved, but are not associated with a droplet.
