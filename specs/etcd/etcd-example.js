var etcd = require("github.com/NetSys/quilt/specs/etcd/etcd");

var nWorker = 3;

var deployment = createDeployment({});

var baseMachine = new Machine({
    provider: "Amazon",
    sshKeys: githubKeys("ejj"), // Replace with your GitHub username.
});

deployment.deploy(baseMachine.asMaster())
deployment.deploy(baseMachine.asWorker().replicate(nWorker + 1))
deployment.deploy(new etcd.Etcd(nWorker));
