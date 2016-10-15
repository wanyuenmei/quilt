var Mongo = require("github.com/NetSys/quilt/specs/mongo/mongo");
var nWorker = 3;

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
var mongo = new Mongo(nWorker);

namespace.deploy(baseMachine.asMaster());
namespace.deploy(baseMachine.asWorker().replicate(nWorker));
namespace.deploy(mongo);
