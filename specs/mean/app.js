var image = "quilt/mean-service";

function App(nWorker, port, env) {
  this._port = port || 80;

  var containers = new Container(image).withEnv(env).replicate(nWorker);
  this._app = new Service("app", containers);

  var hosts = this._app.children();
  for (var i = 0; i < containers.length; i++) {
    containers[i].setEnv("HOST", hosts[i]);
  }

  this._app.connect(this._port, this._app);
}

App.prototype.deploy = function(deployment) {
  deployment.deploy([this._app]);
};

App.prototype.services = function() {
  return [this._app];
};

App.prototype.connect = function(port, to) {
  to.services().forEach(function(service) {
    this._app.connect(port, service);
  }.bind(this));
};

module.exports = App;
