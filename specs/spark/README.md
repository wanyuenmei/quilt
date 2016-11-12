# Quilt: Spark
This document describes how to run Apache Spark on Quilt and pass it a job at
boot time. Specifically, we'll be running the SparkPI example to calculate the
value of π on our Quilt Spark cluster.

## Configuring QUILT_PATH
Quilt uses the `QUILT_PATH` environment variable to locate packages. See the
[Stitch language spec](../../docs/Stitch.md#quilt_path) for instructions on
setting up your `QUILT_PATH`.

To fetch the specs that come with `quilt`, execute `quilt get github.com/NetSys/quilt`.

## SparkPi
The example SparkPi program distributes the computationally-intensive task of
calculating Pi over several machines in a computer cluster.

Our [sparkPI.js](sparkPI.js) Quilt.js specification simplifies the
task of setting up the infrastructure required to run this Spark job.

### Configure SSH authentication
Quilt-managed Machines use public key authentication to control SSH access.
To read the result of the Spark job, we will need to access the Master VM.

If you would like to use `githubKey` authentication, open
`specs/spark/sparkPI.js` and set the `sshKeys` Machine property appropriately:

```javascript
var baseMachine = new Machine({
    ...
    sshKeys: githubKeys("<YOUR_GITHUB_USERNAME>"),
    ...
});
```

For instructions on configuring a user-supplied public key and more information
on configuring Quilt SSH authentication, see
[GettingStarted.md](../../docs/GettingStarted.md#set-up-your-ssh-authentication).

### Build `sparkPI.js`
Start the Quilt daemon with `quilt daemon`. Then, in a separate shell, execute
`quilt run github.com/NetSys/quilt/specs/spark/sparkPI.js` to
build this Quilt.js specification.

Quilt will now begin provisioning several VMs on your cloud provider. Five VMs
will serve as Workers, and one will be the Master.

It will take a bit for the VMs to boot up, for Quilt to configure the network,
and for Docker containers to be initialized. The following output reports that
the Master's public IP is `54.153.115.119`:
```
INFO [Nov 12 09:01:57.191] Successfully booted machines.
...
INFO [Nov 12 09:02:43.819] db.Machine:
	Machine-2{Master, Amazon us-west-1 m4.large, sir-cw7genph, PublicIP=54.153.115.119, PrivateIP=172.31.5.94, Disk=32GB}
	Machine-3{Worker, Amazon us-west-1 m4.large, sir-2i78dmbg, PublicIP=52.53.167.190, PrivateIP=172.31.6.253, Disk=32GB}
...
```

When the `Connected` tag is attached to a Machine description in the console
output, the corresponding VM is fully booted and has began communicating
with Quilt:

```
INFO [Nov 12 09:03:41.606] db.Machine:
	Machine-2{Master, Amazon us-west-1 m4.large, sir-cw7genph, PublicIP=54.153.115.119, PrivateIP=172.31.5.94, Disk=32GB, Connected}
```

Once you see the `Connected` tag, you can connect to the Machine with the
command `quilt ssh <MACHINE_NUM>`, or manually with `ssh quilt@<PUBLIC_IP>`.
That is, in this case either `quilt ssh 2` or `ssh quilt@54.153.115.119`.

### Inspect Containers
To list all active containers in the cluster, use `quilt containers`.  For example:
```
$ quilt containers
Container-9{run quilt/spark run master, Minion: 172.31.6.253, StitchID: 2, IP: 10.14.11.2, Mac: 02:00:0a:0e:0b:02, Labels: [spark-ms], Env: map[JOB:run-example SparkPi]}
Container-10{run quilt/spark run worker, Minion: 172.31.14.240, StitchID: 5, IP: 10.205.149.88, Mac: 02:00:0a:cd:95:58, Labels: [spark-wk], Env: map[MASTERS:1.spark-ms.q]}
Container-11{run quilt/spark run worker, Minion: 172.31.12.226, StitchID: 6, IP: 10.181.3.12, Mac: 02:00:0a:b5:03:0c, Labels: [spark-wk], Env: map[MASTERS:1.spark-ms.q]}
Container-12{run quilt/spark run worker, Minion: 172.31.0.12, StitchID: 7, IP: 10.211.166.114, Mac: 02:00:0a:d3:a6:72, Labels: [spark-wk], Env: map[MASTERS:1.spark-ms.q]}
```

### Recovering Pi
Once our Master Spark container is up, we can find the results of our SparkPi
job via logs.

Execute `quilt logs <Stitch ID>` with the Spark Master node's Stitch ID as
retrieved from the above container descriptions. After scrolling through Spark's
info logging, we will find the result of SparkPi:

```
$ quilt logs 2
# ...
16/06/08 18:49:42 INFO TaskSchedulerImpl: Removed TaskSet 0.0, whose tasks have all completed, from pool
16/06/08 18:49:42 INFO DAGScheduler: ResultStage 0 (reduce at SparkPi.scala:36) finished in 0.381 s
16/06/08 18:49:42 INFO DAGScheduler: Job 0 finished: reduce at SparkPi.scala:36, took 0.525937 s
Pi is roughly 3.13918
16/06/08 18:49:42 INFO SparkUI: Stopped Spark web UI at http://10.0.254.144:4040
16/06/08 18:49:42 INFO MapOutputTrackerMasterEndpoint: MapOutputTrackerMasterEndpoint stopped!
16/06/08 18:49:42 INFO MemoryStore: MemoryStore cleared
16/06/08 18:49:42 INFO BlockManager: BlockManager stopped
```

**Note:** The Spark cluster is now up and usable. You can run the interactive
spark-shell by exec-ing it in the Master Spark container:
`quilt exec <MASTER_CONTAINER_ID> spark-shell`. To tear down the deployment,
just execute `quilt stop`.
