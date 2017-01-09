// AWS: http://docs.aws.amazon.com/cli/latest/reference/ec2/allocate-address.html
// Google: https://cloud.google.com/compute/docs/configure-instance-ip-addresses#reserve_new_static
var floatingIp = "xxx.xxx.xxx.xxx (CHANGE ME)"
var deployment = createDeployment({
  adminACL: ["local"]
});

var webTier = new Service("web_tier", [new Container("nginx:1.10")]);
publicInternet.connect(80, webTier);

webTier.place(new MachineRule(false, {floatingIp: floatingIp}));
deployment.deploy(webTier);

var baseMachine = new Machine({
  provider: "Amazon",
  size: "m4.large",
  region: "us-west-2",
  sshKeys: githubKeys("ejj")
});

deployment.deploy(baseMachine.asMaster());

baseMachine.floatingIp = floatingIp;
deployment.deploy(baseMachine.asWorker());
