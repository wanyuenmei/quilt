var image = "quilt/haproxy"
var cfg = "/usr/local/etc/haproxy/haproxy.cfg";

function Haproxy(n, hosts) {
    var hostnames = hosts.children().join(",");
    var hapRef = new Container(image, [hostnames, "haproxy", "-f", cfg]);
    this.service = new Service("hap", hapRef.replicate(n));
    this.service.connect(80, hosts);

    this.public = function() {
        publicInternet.connect(80, this.service);
    }

    this.deploy = function(deployment) {
        deployment.deploy(this.service);
    };
};

module.exports.Haproxy = Haproxy;
