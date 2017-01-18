var app = require("./app");

var deployment = createDeployment({});

// Setup the infrastructure.
var baseMachine = new Machine({
    provider: "Amazon",
    // Replace with your GitHub username to allow logging into machines.
    // sshKeys: githubKeys("CHANGE_ME"),
});

// Create Master and Worker Machines.
deployment.deploy(baseMachine.asMaster())
deployment.deploy(baseMachine.asWorker().replicate(1));

deployment.deploy(app.New());
