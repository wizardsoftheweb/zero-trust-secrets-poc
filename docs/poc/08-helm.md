# 08 Building a Helm Chart

At this point I'm probably beating a dead horse. I'm trying to poke at all of the things I can think of that might use ZTS. I could make it better or I could make a Helm chart.

## The Chart

I cleaned up the manifests and made things a bit more programmatic. I thought about splitting it into subcharts but that seemed like a lot of work. Sprig doesn't offer much in the way of scripting on top of templating (AFAIK - if I'm wrong I'd love to know) so it's frustrating to mess around with. I guess I'm spoiled by Jinja.

## Improvements

### DockerHub Images

Both containers are in DockerHub now so you can play with them. They're not connected to a pipeline; I was just building and pushing locally.

### `initContainers`

I noticed that, when the chart was first applied, things would go crazy for a bit. Since I slapped the server and client together really fast, they don't do any recovery and just die when they encounter a problem. I prolonged that inevitable refactor by building two `initContainers`.

1. The control server first waits for the `/v2/members` endpoint to respond to requests.
    ```bash
    wget -q -O - http://etcd-client.zts-poc.svc:2379/v2/members
    ``` 
    Once that's up, it snags all the `clientURLs`.
    ```shell-session
    $ wget -q -O - http://etcd-client.zts-poc.svc:2379/v2/members | jq
    {
      "members": [
        {
          "id": "68556d3ade71a836",
          "name": "etcd1",
          "peerURLs": [
            "http://etcd1.zts-poc.svc:2380"
          ],
          "clientURLs": [
            "http://etcd1.zts-poc.svc:2379"
          ]
        },
        {
          "id": "a7a7eda145830594",
          "name": "etcd0",
          "peerURLs": [
            "http://etcd0.zts-poc.svc:2380"
          ],
          "clientURLs": [
            "http://etcd0.zts-poc.svc:2379"
          ]
        }
      ]
    }
    ```
    It waits for each `clientURLs` in order to report good health.
    ```shell-session
    $ wget -q -O - http://etcd1.zts-poc.svc:2379/health | jq
    {
      "health": "true"
    }
    ```
    That might be a little naive, but I was having trouble with Viper not booting until all the nodes were up. I'll tweak it as I go.
    
2. The client needs both the `etcd` hosts and the control server. Since the control server is already waiting on the `etcd` boxes, the clients can just wait on the control server. Once it's up, it should return `pong` on the `/ping` endpoint.
    ```shell-session
    $ wget -q -O - http://control-server.zts-poc.svc:8080/ping | jq
    {
      "message": "pong"
    }
    ```
    
### Larger Deployments

I made it possible to deploy a variable amount of both `etcd` nodes and clients. It was an interesting exercise. It feels weird not increasing the number of replicas in the deployment. However, [the `etcd` manifest I used](https://github.com/etcd-io/etcd/blob/master/hack/kubernetes-deploy/etcd.yml) splits out each host into its own pod and service with all of them sharing the `etcd-client` service. Similarly, the simple clients are supposed to be pretend apps from all sorts of different sources. Also since each client is generated its own key on boot it can't be replicated. Yet.

## Settings

```yaml
# I genuinely don't know why Helm does this
nameOverride: ""
# Ditto
fullnameOverride: ""

# Settings for the etcd deployment(s) and services
etcd:
  # number of nodes you want in the `etcd` cluster
  clusterNodeCount: 2
  # number of replicas of each node
  replicaCount: 1
  # the port for peer URLs
  peerPort: 2380
  # the port for client URLs
  clientPort: 2379

# Settings for the control server deployment and service
controlServer:
  # number of control replicas
  replicaCount: 1
  # the port exposed to the cluster via the service
  clusterPort: 8080
  # the port on each container; containerPort
  targetPort: 8080

# Setting for the simple client deployment(s) and service(s)
simpleClient:
  # number of different clients you'd like to spin up
  desiredCount: 1
  # number of replicas per client
  replicaCount: 1
  # the port exposed to the cluster via the service
  clusterPort: 4747
  # the port on each container; containerPort
  targetPort: 4747
```
