var image = "quilt/memcached";

function Memcached(n) {
    this.service = new Service("memcd", new Container(image).replicate(n));

    this.deploy = function(deployment) {
        deployment.deploy(this.service);
    };
}

module.exports.Memcached = Memcached;
