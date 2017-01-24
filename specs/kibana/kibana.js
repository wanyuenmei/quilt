function Kibana(es) {
    this.service = new Service("kibana", [
        new Container("kibana:4",
            [
                "--port", this.port.toString(),
                "--elasticsearch", es.uri(),
            ]
        )
    ]);
    es.addClient(this.service);
}

Kibana.prototype.public = function() {
    publicInternet.connect(this.port, this.service);
    return this;
}

Kibana.prototype.deploy = function(depl) {
    depl.deploy(this.service);
}

Kibana.prototype.port = 5601;

exports.Kibana = Kibana;
