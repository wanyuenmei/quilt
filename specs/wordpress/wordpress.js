var image = "quilt/wordpress";

function Wordpress(n, db, memcd) {
    var refContainer = new Container(image)
        .withEnv({
            "MEMCACHED": memcd.service.children().join(","),
            "DB_MASTER": db.master.children().join(","),
            "DB_REPLICA": db.replicas.children().join(","),
        });
    this.wp = new Service("wp", refContainer.replicate(n));
    this.wp.connect(3306, db.replicas);
    this.wp.connect(3306, db.master);
    this.wp.connect(11211, memcd.service);

    this.deploy = function(deployment) {
        deployment.deploy(this.wp);
    };
}

module.exports.Wordpress = Wordpress;
