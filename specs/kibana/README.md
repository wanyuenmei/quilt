# Elasticsearch and Kibana

[main.js](main.js) contains an example deployment that boots an Elasticsearch
cluster connected to Kibana, and makes the Kibana web interface publically
accessible.

[Elasticsearch](https://www.elastic.co/products/elasticsearch) is a
"distributed, RESTful search and analytics engine".

[Kibana](https://www.elastic.co/products/kibana) "lets you visualize your
Elasticsearch data".

## Booting The Deployment

This guide assumes Quilt is installed on your machine. If you have not yet
setup Quilt, please follow the instructions in our [getting started
guide](../../docs/GettingStarted.md). Specifically, you need to have:
- The Quilt binary installed
- The Quilt daemon running
- AWS credentials setup
- `QUILT_PATH` configured

To deploy the Elasticsearch and Kibana cluster, simply `quilt run main.js`.

You can track the status of the deployment with `quilt ps`. Once the status of
all containers is `Running`, we can begin interacting with our cluster.

## Interacting With The Cluster

Quilt's job is basically done at this point: the containers are up, and can
communicate with each other over the network. If a container happens to die,
Quilt will handle restarting it for you.

However, we'll still walk through [elastic's
example](https://www.elastic.co/guide/en/kibana/3.0/using-kibana-for-the-first-time.html)
of analyzing Shakespeare's works to show off some of the tools that come with
Quilt.

### Seeding Elasticsearch

Our current deployment doesn't allow public access to Elasticsearch, but we
need to access it to seed it with data. To do so, modify [main.js](main.js) as
follows:

```
-var es = new Elasticsearch(clusterSize);
+var es = new Elasticsearch(clusterSize).public();
```

Save your change, then `quilt run main.js` for the change to be deployed. The
diff showed by `quilt run` should show that `elasticsearch` now accepts
connections from `public`.

```
+ "From": "public",
+ "To": "elasticsearch",
+ "MinPort": 9200,
+ "MaxPort": 9200
```

To get the public IP of an Elasticsearch container, run `quilt containers` and
copy the `PUBLIC IP` of any of the Elasticsearch containers. It does not matter
which Elasticsearch container is used -- interacting with any of them will
propagate changes to the entire cluster.

```
% quilt containers
ID    MACHINE      CONTAINER            LABELS           STATUS     PUBLIC IP
2     Machine-3    elasticsearch:2.4    elasticsearch    Running    54.153.44.229:9200
3     Machine-4    elasticsearch:2.4    elasticsearch    Running    52.53.172.21:9200
4     Machine-4    kibana:4             kibana           Running    52.53.172.21:5601
```

As a sanity check, make sure the container responds.

```
% curl 54.153.44.229:9200
{
  "name" : "Johann Schmidt",
  "cluster_name" : "elasticsearch",
  "cluster_uuid" : "6uh_et8iR2GcFfHuFEr06A",
  "version" : {
    "number" : "2.4.4",
    "build_hash" : "fcbb46dfd45562a9cf00c604b30849a6dec6b017",
    "build_timestamp" : "2017-01-03T11:33:16Z",
    "build_snapshot" : false,
    "lucene_version" : "5.5.2"
  },
  "tagline" : "You Know, for Search"
}
```

Now, create the index.

```
% curl -XPUT http://54.153.44.229:9200/shakespeare -d '
{
 "mappings" : {
  "_default_" : {
   "properties" : {
    "speaker" : {"type": "string", "index" : "not_analyzed" },
    "play_name" : {"type": "string", "index" : "not_analyzed" },
    "line_id" : { "type" : "integer" },
    "speech_number" : { "type" : "integer" }
   }
  }
 }
}
';
```

And populate it.

```
curl https://www.elastic.co/guide/en/kibana/3.0/snippets/shakespeare.json | curl -XPUT 54.153.44.229:9200/_bulk --data-binary @-
```

### Using Kibana

With data in Elasticsearch, we can now view it in Kibana. We can find the
public IP of the Kibana container using `quilt containers` as before.

Go to the address (`54.183.201.132:5601` for the above output) in your browser.
Uncheck `Index contains time-based events`, set the index to be `shakes*`, and
hit `Create`.

We can now issue queries and create dashboards to analyze Shakespeare's works.
For example, to find lines containing the words "friends", "Romans", or
"countrymen", click `Discover` in the navbar, and search `friends, Romans,
countrymen`.

## Cleaning Up

Once done with the cluster, don't forget to destroy it with `quilt stop`.
