var image = "quilt/haproxy";
var cfg = "/usr/local/etc/haproxy/haproxy.cfg";

function Haproxy(n, services, port) {
    services = Array.isArray(services) ? services : [services];
    port = port || 80;

    var hostnames = _.flatten(services.map(function(service) {
      return service.children();
    }));
    var addresses = hostnames.map(function(host) {
      return host + ":" + port;
    });
    var hapRef = new Container(image).withEnv({
      "ADDRS": addresses.join(",")
    });

    this.service = new Service("hap", hapRef.replicate(n));
    services.forEach(function(service) {
      this.service.connect(port, service);
    }.bind(this));

    this.public = function() {
      publicInternet.connect(80, this.service);
    };

    this.deploy = function(deployment) {
        deployment.deploy(this.service);
    };
};

module.exports.Haproxy = Haproxy;
