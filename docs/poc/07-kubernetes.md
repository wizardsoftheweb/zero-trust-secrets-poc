# 07 Kubernetes

You'll want [`kubectx`/`kubens`](https://github.com/ahmetb/kubectx).

These are assumed to be run before anything else.
```shell-session
cd path/to/kubernetes
minikube start
kubectl config use-context minikube
eval $(minikube docker-env)
alias k=kubectl
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
$ k apply -f pod_etcd0.yaml
pod/etcd0 created

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
