function App(cfg) {
  if (typeof cfg.nWorker !== 'number') {
    throw new Error('`nWorker` is required');
  }
  if (typeof cfg.image !== 'string') {
    throw new Error('`image` is required');
  }

  this._port = cfg.port || 80;

  var env = cfg.env || {};
  var containers = new Container(cfg.image).withEnv(env).replicate(cfg.nWorker);
  this._app = new Service("app", containers);

  var hosts = this._app.children();
  for (var i = 0; i < containers.length; i++) {
    containers[i].setEnv("HOST", hosts[i]);
  }

  this._app.connect(this._port, this._app);
};

App.prototype.deploy = function(deployment) {
  deployment.deploy([this._app]);
};

App.prototype.services = function() {
  return [this._app];
};

App.prototype.port = function() {
  return this._port;
};

App.prototype.connect = function(port, to) {
  to.services().forEach(function(service) {
    this._app.connect(port, service);
  }.bind(this));
};

module.exports = App;
