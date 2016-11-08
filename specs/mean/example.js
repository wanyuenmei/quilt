var HaProxy = require("github.com/NetSys/quilt/specs/haproxy/haproxy").Haproxy;
var Mongo = require("github.com/NetSys/quilt/specs/mongo/mongo");
var App = require("github.com/NetSys/quilt/specs/mean/app");

// AWS
var namespace = createDeployment({
    namespace: "CHANGE_ME",
    adminACL: ["local"],
});
var baseMachine = new Machine({
    provider: "Amazon",
    cpu: new Range(2),
    ram: new Range(2),
    sshKeys: githubKeys("ejj"),
});
namespace.deploy(baseMachine.asMaster());
namespace.deploy(baseMachine.asWorker().replicate(3));

var mongo = new Mongo(3);
var app = new App({
  nWorker: 3,
  port: 8080,
  env: {
    MONGO_URI: mongo.uri("mean-example")
  }
});
var haproxy = new HaProxy(3, app.services(), 8080);

// Places all haproxy containers on separate Worker VMs.
// This is just for convenience for the example instructions, as it allows us to
// access the web application by using the IP address of any Worker VM.
haproxy.service.place(new LabelRule(true, haproxy.service));

mongo.connect(mongo.port(), app);
app.connect(mongo.port(), mongo);
haproxy.public();

namespace.deploy(app);
namespace.deploy(mongo);
namespace.deploy(haproxy);
