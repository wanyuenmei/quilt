var image = "quilt/mysql";

function Mysql(n) {
    this.master = new Service("mysql-dbm",
            [new Container(image, ["--master", "1", "mysqld"])]);

    var replicas = [];
    var mHost = this.master.children().join(",");
    for (var i = 2; i < (n + 2); i++) {
        replicas.push(new Container(image, ["--replica", mHost, i.toString(), "mysqld"]));
    }
    this.replicas = new Service("mysql-dbr", replicas);

    this.replicas.connect(3306, this.master);
    this.replicas.connect(22, this.master);

    this.deploy = function(deployment) {
        deployment.deploy(this.master);
        deployment.deploy(this.replicas);
    };
}

module.exports.Mysql = Mysql;
