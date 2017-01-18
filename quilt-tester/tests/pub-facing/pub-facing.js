var nginx = require("github.com/NetSys/quilt/specs/nginx/app");
var infrastructure = require("github.com/NetSys/quilt/quilt-tester/config/infrastructure")

var deployment = createDeployment({});
deployment.deploy(infrastructure);

var nWorker = deployment.machines.filter(function(m) {
    return m.role == "Worker"
}).length;

for (var i = 0 ; i < nWorker ; i++) {
    deployment.deploy(nginx.New(80));
    deployment.deploy(nginx.New(8000));
}
