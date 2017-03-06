var infrastructure = require("github.com/quilt/quilt/quilt-tester/config/infrastructure")

var deployment = createDeployment();
deployment.deploy(infrastructure);

var nWorker = deployment.machines.filter(function(m) {
    return m.role == "Worker"
}).length;

for (var i = 0 ; i < nWorker ; i++){
    deployment.deploy(new Service("foo",
        new Container(
            new Image("test-custom-image" + i,
                "FROM alpine\n" +
                "RUN echo " + i + " > /dockerfile-id\n" +
                "RUN echo $(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1) > /image-id"
            ),
            ["tail", "-f", "/dev/null"]
        ).replicate(2)
    ));
}
