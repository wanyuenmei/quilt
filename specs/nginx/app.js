var image = "quilt/nginx"

exports.New = function(port) {
    port = port || 80;
    if (typeof port !== 'number') {
        throw new Error("port must be a number");
    }

    // Create a Nginx Docker container, encapsulating it within the service "web_tier".
    var webTier = new Service("web_tier", [
            new Container(image).withEnv({
                PORT: port.toString(),
            })
    ]);
    publicInternet.connect(port, webTier);

    return webTier;
}
