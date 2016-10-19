// Import spark.spec
var spark = require("github.com/NetSys/quilt/specs/spark/spark");

var deployment = createDeployment({
    // Using unique Namespaces will allow multiple Quilt instances to run on the
    // same cloud provider account without conflict.
    namespace: "CHANGE_ME",
    // Defines the set of addresses that are allowed to access Quilt VMs.
    adminACL: ["local"],
});

// We will have three worker machines.
var nWorker = 3;

// Application
// sprk.exclusive enforces that no two Spark containers should be on the
// same node. sprk.public says that the containers should be allowed to talk
// on the public internet. sprk.job causes Spark to run that job when it
// boots.
var sprk = new spark.Spark(1, nWorker)
    .exclusive()
    .public()
    .job("run-example SparkPi");

// Infrastructure
var baseMachine = new Machine({
    provider: "Amazon",
    region: "us-west-1",
    size: "m4.large",
    diskSize: 32,
    sshKeys: githubKeys("ejj"), // Replace with your GitHub username.
});
deployment.deploy(baseMachine.asMaster())
deployment.deploy(baseMachine.asWorker().replicate(nWorker + 1))
deployment.deploy(sprk);

deployment.assert(publicInternet.canReach(sprk.masters), true);
deployment.assert(enough, true);
