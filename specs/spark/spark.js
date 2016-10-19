var image = "quilt/spark";

function setImage(newImage) {
    image = newImage;
}

function Spark(nMaster, nWorker, zookeeper) {
    var dkms = new Container(image, ["run", "master"]).replicate(nMaster);

    if (zookeeper) {
        var zooHosts = zookeeper.children().join(",");
        for (var i = 0; i < nMaster; i++) {
            dkms[i].setEnv("ZOO", zooHosts);
        }
    }

    this.masters = new Service("spark-ms", dkms);

    var dkws = new Container(image, ["run", "worker"])
        .withEnv({"MASTERS": this.masters.children().join(",")})
        .replicate(nWorker);
    this.workers = new Service("spark-wk", dkws);

    this.workers.connect(7077, this.workers);
    this.workers.connect(7077, this.masters);
    if (zookeeper) {
        this.masters.connect(2181, zookeeper);
    }

    this.job = function(command) {
        var cnt = this.masters.containers;
        for (var i = 0; i < cnt.length; i++) {
            cnt[i].env["JOB"] = command;
        }
        return this;
    }

    this.public = function() {
        publicInternet.connect(8080, this.masters);
        publicInternet.connect(8081, this.workers);
        return this;
    }

    this.exclusive = function() {
        this.masters.place(new LabelRule(true, this.workers));
        return this;
    }

    this.deploy = function(deployment) {
        deployment.deploy(this.masters);
        deployment.deploy(this.workers);
    }
}

exports.setImage = setImage;
exports.Spark = Spark;
