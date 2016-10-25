// Create a new deployment.
// Using unique Namespaces will allow multiple Quilt instances to run on the
// same cloud provider account without conflict.
// Also defines the set of addresses that are allowed to access Quilt VMs.
var deployment = createDeployment({
    namespace: "CHANGE_ME",
    adminACL: ["local"],
});

// Create a Nginx Docker container, encapsulating it within the service "web_tier".
var webTier = new Service("web_tier", [new Container("nginx:1.10")]);
publicInternet.connect(80, webTier);

// Services must be explicitly deployed.
deployment.deploy(webTier);

// Setup the infrastructure.
var baseMachine = new Machine({
    provider: "Amazon",
    size: "m4.large",
    sshKeys: githubKeys("ejj"), // Replace with your GitHub username.
});

// Create Master and Worker Machines.
deployment.deploy(baseMachine.asMaster())
deployment.deploy(baseMachine.asWorker().replicate(2));
