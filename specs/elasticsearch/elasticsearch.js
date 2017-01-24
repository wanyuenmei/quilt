function Elasticsearch(n) {
    var ref = new Container("elasticsearch:2.4",
        [
            "--transport.tcp.port", this._transportPorts.min + "-" + this._transportPorts.max,
            "--http.port", this.port.toString(),
        ]
    )
    this.service = new Service("elasticsearch", ref.replicate(n));

    if (n > 1) {
        var hosts = this.service.children();
        var hostsStr = hosts.join(",");
        this.service.containers.forEach(function(c, i) {
            c.command.push(
                "--discovery.zen.ping.unicast.hosts", hostsStr,
                "--network.host", hosts[i]
            );
        });
    }

    this.service.connect(this._transportPorts, this.service);
}

Elasticsearch.prototype.uri = function() {
    return "http://" + this.service.hostname() + ":" + this.port;
}

Elasticsearch.prototype.public = function() {
    publicInternet.connect(this.port, this.service);
    return this;
}

Elasticsearch.prototype.addClient = function(clnt) {
    clnt.connect(this.port, this.service);
}

Elasticsearch.prototype.deploy = function(depl) {
    depl.deploy(this.service);
}

Elasticsearch.prototype._transportPorts = new PortRange(9300, 9400);
Elasticsearch.prototype.port = 9200;

exports.Elasticsearch = Elasticsearch;
