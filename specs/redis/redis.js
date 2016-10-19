var image = "quilt/redis";

function Redis(nWorker, auth) {
    this.master = createMaster(auth);
    this.workers = createWorkers(nWorker, auth, this.master);
    this.master.connect(6379, this.workers);
    this.workers.connect(6379, this.master);

    this.exclusive = function() {
        this.master.place(new LabelRule(true, this.workers));
    }

    this.deploy = function(deployment) {
        deployment.deploy([this.master, this.workers]);
    }
}

function createMaster(auth) {
    return new Service("redis-ms", [
        new Container(image, ["run"]).withEnv({
            "ROLE": "master",
            "AUTH": auth,
        })
    ]);
}

function createWorkers(n, auth, master) {
    var refWorker = new Container(image, ["run"]).withEnv({
        "ROLE": "worker",
        "AUTH": auth,
        "MASTER": master.hostname(),
    });
    return new Service("redis-wk", refWorker.replicate(n));
}

module.exports.Redis = Redis;
