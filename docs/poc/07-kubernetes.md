# 07 Kubernetes

You'll want [`kubectx`/`kubens`](https://github.com/ahmetb/kubectx).

These are assumed to be run before anything else.
```shell-session
cd path/to/kubernetes
minikube start
kubectl config use-context minikube
eval $(minikube docker-env)
alias k=kubectl
docker run -d -p 5000:5000 --restart=always --name registry registry:2 
```

## Namespace

```shell-session
$ k apply -f namespace_zts-poc.yaml
namespace/zts-poc created

$ kubens zts-poc
Context "minikube" modified.
Active namespace is "zts-poc".
```

## `etcd`

I borrowed a very simple deployment from [the `etcd` repo](https://github.com/etcd-io/etcd/blob/master/hack/kubernetes-deploy/etcd.yml).

```shell-session
$ k apply -f deployment_etcd0.yaml
deployment.apps/etcd0 created

$ k apply -f service_etcd0.yaml
service/etcd0 created

$ k apply -f service_etcd-client.yaml
service/etcd-client created

$ k get po
NAME    READY   STATUS    RESTARTS   AGE
etcd0   1/1     Running   0          53s

$ k get svc
NAME          TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
etcd-client   NodePort   10.96.124.171   <none>        2379:30646/TCP                  16m
etcd0         NodePort   10.97.161.16    <none>        2379:31823/TCP,2380:30421/TCP   16m


$ minikube service --namespace zts-poc list
|-----------|-------------|--------------------------------|
| NAMESPACE |    NAME     |              URL               |
|-----------|-------------|--------------------------------|
| zts-poc   | etcd-client | http://192.168.99.101:30646    |
| zts-poc   | etcd0       | http://192.168.99.101:31823    |
|           |             | http://192.168.99.101:30421    |
|-----------|-------------|--------------------------------|

# This is going to fail b/c it's run on the box running Minikube
# Outside of Minikube there's no resolution on etcd0
$ etcdctl --endpoint http://192.168.99.101:30646 --no-sync cluster-health
failed to check the health of member cf1d15c5d194b5c9 on http://etcd0:2379: Get http://etcd0:2379/health: dial tcp 23.202.231.166:2379: connect: connection refused
member cf1d15c5d194b5c9 is unreachable: [http://etcd0:2379] are all unreachable
cluster is unhealthy

$ etcdctl --endpoints 'http://192.168.99.101:30646' --no-sync ls / 


$ etcdctl --endpoints 'http://192.168.99.101:30646' --no-sync mkdir /test

$ etcdctl --endpoints 'http://192.168.99.101:30646' --no-sync ls / 
/test

$ etcdctl --endpoints 'http://192.168.99.101:30646' --no-sync rmdir /test 
```

## Control Server

The first thing to do is get the image in the registry we set up. If you're having trouble, [this gist](https://gist.github.com/kevin-smets/b91a34cea662d0c523968472a81788f7) was fantastic. You might be missing a dependency.
```shell-session
cd path/to/control-server
docker build -t control-server .
docker tag control-server localhost:5000:1.0.0
```

With that out the way, we can run kube stuff.
```shell-session
$ cd path/to/kubernetes
$ k apply -f deployment_control-server.yaml
deployment.apps/control-server created

$ k apply -f service_control-server.yaml 
service/control-server created

$ k get svc
NAME             TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
control-server   NodePort   10.106.82.166   <none>        8080:30932/TCP                  5m57s
etcd-client      NodePort   10.96.124.171   <none>        2379:30646/TCP                  60m
etcd0            NodePort   10.97.161.16    <none>        2379:31823/TCP,2380:30421/TCP   61m

$ minikube service --namespace zts-poc list 
|-----------|----------------|--------------------------------|
| NAMESPACE |      NAME      |              URL               |
|-----------|----------------|--------------------------------|
| zts-poc   | control-server | http://192.168.99.101:30932    |
| zts-poc   | etcd-client    | http://192.168.99.101:30646    |
| zts-poc   | etcd0          | http://192.168.99.101:31823    |
|           |                | http://192.168.99.101:30421    |
|-----------|----------------|--------------------------------|

$ curl -s 'http://192.168.99.101:30932/ping' | jq
{
  "message": "pong"
}

$ k logs $(k get po | awk '/^control-server/{ print $1; }') 
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /metrics                  --> github.com/zsais/go-gin-prometheus.prometheusHandler.func1 (4 handlers)
[GIN-debug] GET    /ping                     --> main.main.func1 (4 handlers)
[GIN-debug] POST   /rando                    --> main.main.func2 (4 handlers)
[GIN-debug] Environment variable PORT="8080"
[GIN-debug] Listening and serving HTTP on :8080
[GIN] 2019/07/13 - 16:43:55 | 200 |     275.747µs |      172.17.0.1 | GET      /ping
```

## Simple Client

