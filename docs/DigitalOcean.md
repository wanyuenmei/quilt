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

1. Create a new key [here](cloud.digitalocean.com/settings/api/tokens).
Both read and write permissions are required.

2. Save the key in `~/.digitalocean/key` on the machine that will be running the
Quilt daemon.
