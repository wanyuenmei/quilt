var image = "quilt/mongo";
var PORT = 27017;

function Mongo(nWorker) {
  this._port = PORT;

  // The "initial" container is the node the replica set is initialized from.
  // Initializing a replica set triggers an election of primary (master),
  // secondaries (slaves), etc. The initial container is not necessarily the primary.
  var initialContainer = new Container(image);
  initialContainer.setEnv("ROLE", "initial");

  this._initial = new Service("mongo-initial", [initialContainer]);
  initialContainer.setEnv("HOST", this._initial.hostname());

  // "subsequent" containers are peers of the "initial" container. They receive the
  // starting replica set configuration from the initial container.
  var subsequentContainers = new Container(image).replicate(nWorker-1);
  this._subsequents = new Service("mongo-subsequent", subsequentContainers);

  var subsequentHosts = this._subsequents.children();
  for (var i = 0; i < subsequentContainers.length; i++) {
    subsequentContainers[i].setEnv("ROLE", "subsequent");
    subsequentContainers[i].setEnv("HOST", subsequentHosts[i]);
  }

  // `PEERS` tells the "initial" container where to find the other members of the
  // replica set ("subsequent" containers).
  initialContainer.setEnv("PEERS", subsequentHosts.join(","));

  // All containers should be able to communicate in/out, as they could all
  // serve either role (primary, secondary), given a re-election upon failure, etc.
  this._initial.connect(this._port, this._initial);
  this._initial.connect(this._port, this._subsequents);
  this._subsequents.connect(this._port, this._initial);
  this._subsequents.connect(this._port, this._subsequents);
};

Mongo.prototype.deploy = function(deployment) {
  deployment.deploy([this._initial, this._subsequents]);
};

Mongo.prototype.services = function() {
  return [this._initial, this._subsequents];
};

Mongo.prototype.connect = function(p, to) {
  to.services().forEach(function(service) {
    this._initial.connect(p, service);
    this._subsequents.connect(p, service);
  }.bind(this));
};

Mongo.prototype.uri = function(dbName) {
  var services = [this._initial, this._subsequents];
  var hostnames = _.flatten(services.map(function(service) {
    return service.children();
  }));
  return "mongodb://" + hostnames.join(",") + "/" + dbName;
};

Mongo.prototype.port = function() {
  return this._port;
};

module.exports = Mongo;
