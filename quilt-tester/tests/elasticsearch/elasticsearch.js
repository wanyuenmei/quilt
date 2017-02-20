var Elasticsearch = require("github.com/quilt/elasticsearch").Elasticsearch;
var infrastructure = require("github.com/quilt/quilt/quilt-tester/config/infrastructure")

var deployment = createDeployment();
deployment.deploy(infrastructure);

var nWorker = deployment.machines.filter(function(m) {
    return m.role == "Worker"
}).length;
deployment.deploy(new Elasticsearch(nWorker).public());
