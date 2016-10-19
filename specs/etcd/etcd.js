var image = "quilt/etcd";

function Etcd(n) {
    var refContainer = new Container(image);
    this.etcd = new Service("etcd", refContainer.replicate(n));
    var children = this.etcd.children();
    var peers = children.join(",");
    for (var i = 0; i < this.etcd.containers.length; i++) {
        this.etcd.containers[i].setEnv("PEERS", peers);
        this.etcd.containers[i].setEnv("HOST", children[i]);
    }
    this.etcd.connect(new PortRange(1000, 65535), this.etcd);

    this.deploy = function(deployment) {
        deployment.deploy(this.etcd);
    }
}

module.exports.Etcd = Etcd;
