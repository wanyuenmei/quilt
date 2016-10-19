var redis = require("github.com/NetSys/quilt/specs/redis/redis");

var deployment = createDeployment({
    // Using unique Namespaces will allow multiple Quilt instances to run on the
    // same cloud provider account without conflict.
    Namespace: "CHANGE_ME",
});

var nWorker = 1;

// Boot redis with 2 workers and 1 master. AUTH_PASSWORD is used to secure
// the redis connection
var rds = new redis.Redis(nWorker, "AUTH_PASSWORD");
rds.exclusive();

var baseMachine = new Machine({
    provider: "Amazon",
    sshKeys: githubKeys("ejj"), // Replace with your GitHub username.
});

deployment.deploy(baseMachine.asMaster())
deployment.deploy(baseMachine.asWorker().replicate(nWorker + 1))
deployment.deploy(rds);
