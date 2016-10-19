var zookeeper = require("github.com/NetSys/quilt/specs/zookeeper/zookeeper");

var n = 3;
var zoo = new zookeeper.Zookeeper(n);
var deployment = createDeployment({
    namespace: "CHANGE_ME",
    adminACL: ["local"],
});

var baseMachine = new Machine({
    provider: "Amazon",
    region: "us-west-1",
    size: "m4.large",
    diskSize: 32,
    sshKeys: githubKeys("ejj"), // Replace with your GitHub username.
});

deployment.deploy(baseMachine.asMaster())
deployment.deploy(baseMachine.asWorker().replicate(n))
deployment.deploy(zoo);
