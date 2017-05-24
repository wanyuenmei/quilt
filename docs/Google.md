# Google Compute Engine

Quilt supports the `Google` provider for booting instances on the Google Compute
Engine. For example, to deploy a GCE machine in the `us-east1-b` zone of size
`n1-standard-1` as a `Worker`:

```
deployment.deploy(new Machine({
  provider: "Google",
  region: "us-east1-b",
  size: "n1-standard-1",
  role: "Worker"
}));
```

## Setup

1. Create a Google Cloud Platform Project

All instances are booted under a Cloud Platform project. To setup a project for
use with Quilt, go to the [console page](http://console.cloud.google.com), then
click the project dropdown at the top of page, and hit the plus icon. Pick a
name, and create your project.

2. Enable the Compute API

Select your newly created project from the project selector at the top of the
[console page](http://console.cloud.google.com), and then select `API Manager
-> Library` from the navbar on the left. Search for and enable the `Google
Compute Engine API`.

3. Save the Credentials File

Go to `Credentials` on the left navbar (under `API Manager`), and create
credentials for a `Service account key`. Create a new service account with the
`Project -> Editor` role, and select the JSON output option. Copy the
downloaded file to `~/.gce/quilt.json` on the machine from which you will be
running the Quilt daemon.

That's it! You should now be able to boot machines on the `Google` provider.
