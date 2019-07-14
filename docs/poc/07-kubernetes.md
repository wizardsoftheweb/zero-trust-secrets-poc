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
docker tag control-server localhost:5000/control-server:1.0.0
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

Same process with the registry.
```shell-session
cd path/to/simple-client
docker build -t simple-client .
docker tag simple-client localhost:5000/simple-client:1.0.0
```
With that done, we can wire up the last part!
```shell-session
$ cd path/to/kubernetes
$ k apply -f deployment_simple-client.yaml
deployment.apps/simple-client created

$ k apply -f service_simple-client0.yaml
service/simple-client0 created

$ k get svc
NAME             TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)                         AGE
control-server   NodePort   10.106.82.166   <none>        8080:30932/TCP                  67m
etcd-client      NodePort   10.96.124.171   <none>        2379:30646/TCP                  122m
etcd0            NodePort   10.97.161.16    <none>        2379:31823/TCP,2380:30421/TCP   122m
simple-client0   NodePort   10.106.96.139   <none>        4747:30745/TCP                  32m

$ minikube service --namespace zts-poc list
|-----------|----------------|--------------------------------|
| NAMESPACE |      NAME      |              URL               |
|-----------|----------------|--------------------------------|
| zts-poc   | control-server | http://192.168.99.101:30932    |
| zts-poc   | etcd-client    | http://192.168.99.101:30646    |
| zts-poc   | etcd0          | http://192.168.99.101:31823    |
|           |                | http://192.168.99.101:30421    |
| zts-poc   | simple-client0 | http://192.168.99.101:30745    |
|-----------|----------------|--------------------------------|

$ curl -s 'http://192.168.99.101:30745/ping' | jq
{
  "message": "pong"
}

$ curl -s 'http://192.168.99.101:30745/' | jq
{
  "secrets": [
    "SzF-y1CQvy5GdKKPBj47IioY84GE4V1uQkvD3e5f-A4FmWdIquwjU5onQg2VWLo=",
    "v9cJZ3csmISkxB8ymPe85z2FLmiOX7iEPM5ZI_faXaAQJXgdGuSDnSp5-mSynD4=",
    "2wWLFUZjXfcQkZDSitDVElDZPaV0WIxKQR2wXDSD7CvjXWFq_-WYGpForfAkfJg=",
    "dT8B9UfgLbXgMOP_DMR2UVEsV0hHokJC9DjO-W2EVynTyQTQN45zjA7_TJkyxck=",
    "SWM1ZqN1ZUSX3k_Dbcx_XBxvnlBUhDJvqcqyhSStmJlgew7KqLjqF1ctYHERF6Y=",
    "b4F48EkwDStyrFumiN2VtPIAY-Q_gDeETF778TJkEdR9tfc_B5HjM397ztlkehE=",
    "etKXLhchW2oeESpr-2JQN7ADEMvVwYlGBnp3NoCcPt-CxW2qDRcT38qZKfmAuPc=",
    "3WHMG2Q86si-A-h2M5oX79sgZbuj0Q5GSLj4Q4Z_QUWcc10hs0q3d4_PVD4CGog=",
    "rxwLr2_z1nr_i-_BqtutuMBaln2vLKbSRl8liI3n-h033xLyY0icPfOi_61mARU=",
    "XSC4K21hTG8NCGHCcxqhxA1OQP5fmb2WRQczhUOO2REizv6kz6-W47HZsgSEXsI="
  ]
}

$ curl -s 'http://192.168.99.101:30745' | jq; \
    curl -s 'http://192.168.99.101:30745/force-update' | jq; \
    sleep 30; \
    curl -s 'http://192.168.99.101:30745' | jq
{
  "secrets": [
    "SzF-y1CQvy5GdKKPBj47IioY84GE4V1uQkvD3e5f-A4FmWdIquwjU5onQg2VWLo=",
    "v9cJZ3csmISkxB8ymPe85z2FLmiOX7iEPM5ZI_faXaAQJXgdGuSDnSp5-mSynD4=",
    "2wWLFUZjXfcQkZDSitDVElDZPaV0WIxKQR2wXDSD7CvjXWFq_-WYGpForfAkfJg=",
    "dT8B9UfgLbXgMOP_DMR2UVEsV0hHokJC9DjO-W2EVynTyQTQN45zjA7_TJkyxck=",
    "SWM1ZqN1ZUSX3k_Dbcx_XBxvnlBUhDJvqcqyhSStmJlgew7KqLjqF1ctYHERF6Y=",
    "b4F48EkwDStyrFumiN2VtPIAY-Q_gDeETF778TJkEdR9tfc_B5HjM397ztlkehE=",
    "etKXLhchW2oeESpr-2JQN7ADEMvVwYlGBnp3NoCcPt-CxW2qDRcT38qZKfmAuPc=",
    "3WHMG2Q86si-A-h2M5oX79sgZbuj0Q5GSLj4Q4Z_QUWcc10hs0q3d4_PVD4CGog=",
    "rxwLr2_z1nr_i-_BqtutuMBaln2vLKbSRl8liI3n-h033xLyY0icPfOi_61mARU=",
    "XSC4K21hTG8NCGHCcxqhxA1OQP5fmb2WRQczhUOO2REizv6kz6-W47HZsgSEXsI="
  ]
}
{
  "message": "Secrets were regenerated"
}
{
  "secrets": [
    "dN2CaqbKneEFwP9LDTkf8KhFRR-UZnyCHVSo_rbnRY2SvVb41g2iiSipj-BavC8=",
    "Y22eROU0GmcOdWQTW1XfUQLqkjGfrG2XsyBRqBC9j9_t1jqRYUsR47hviynBOl8=",
    "Q1Zhu78aLUG8Zr8t4WqOeHwRRFe3gXoKBo5rr2ArcAEkRRxqZdbK09VJ4rKwtO4=",
    "G_EkSQSvfpdpkAPAvVkUOfUAWWcoAdP72dcpTauAcY-Swjnyitw8SFrTQEor0Xo=",
    "vk6htwcXZHvo3jVFccV6cW9RkAjCew3b-xgs19GDwMrQHQG5rmzspQfY0N8WdOE=",
    "mW3p5yw0WmEh4VUz5LTki6_BuJYakTn19b3gs_vx9Br9HmfLMqkP3WHeWy01lGQ=",
    "Q35Xc4XhEGlaMxe3CS1W21XvBgH3MpA-zrYLvmisGdXvK8FSjpIqcNEr9Dir5iM=",
    "i7GQJrXm4n3TWL4WYFAYcjNs6NoSrHGBDBC8aMZYNjx3Q4rzH0tj0sUFAWanSh0=",
    "oFpUa7TzMPAmcJtq12gpvyGXjl1pHm0PXnoJm6Ti2m3w6eboJX0S7xjKRjMNxq0=",
    "HqGpI8ll8Y-qC35kOUieHuoaaepp9nwrokmTylR073Yp60GLHwjQlq69jCtrZws="
  ]
}

$ k logs $(k get po | awk '/^simple-client/{ print $1; }')
/tmp/509691206
gpg: Generating a configuration OpenPGP key
gpg: key 80293F442B84822C marked as ultimately trusted
gpg: directory '/appuser/.gnupg/openpgp-revocs.d' created
gpg: revocation certificate stored as '/appuser/.gnupg/openpgp-revocs.d/FB9A2B6684C5BD25226400A080293F442B84822C.rev'
gpg: done

2019/07/13 17:43:03 &main.RandoRequest{Count:10, KvHosts:[]string{"http://etcd0.zts-poc.svc:2379/"}, KvKey:"/simple-client0/secrets.json", PubKey:"-----BEGIN PGP PUBLIC KEY BLOCK-----\\n\\nmQENBF0qGCYBCADLjTH3IqfcVs0X/x4xN+7AT+6cgvy7sYNHMCDbo2h/KCVRJo6c\\nmqw44gbTktzxzs3SugJFVIaWpsSsYO+VcmSvmufojnNmP5Sk4YVcM4MSXs0BouN1\\noCSw8gAzWh8rGEd5suVTzyX6YFUVO6RrjxVv9VkuFTmfotgQyEtxcO83YgqCZaqI\\nIlSej5Cfnqx4ccoH4fBISz+B07aINRZy3H5F27QGbG3g26hhcVo8GChy0hwl4psQ\\ns3AEJpWIhHFZpHfMXZkEWK1+JbbBS7aerDxJFDwA1HgnhhmWmuKEB70FhCl8bw0K\\n4OnSzlqS0kkfimUd5FqnNacaEEmzQk5+1JHzABEBAAG0LUNKIEhhcnJpZXMgKFpl\\ncm8gVHJ1c3QgU2VjcmV0cykgPGNqQHdvdHcucHJvPokBTgQTAQgAOBYhBPuaK2aE\\nxb0lImQAoIApP0QrhIIsBQJdKhgmAhsDBQsJCAcCBhUKCQgLAgQWAgMBAh4BAheA\\nAAoJEIApP0QrhIIsot8H/jtTV8ndJA83iB0Cf2oXnRFEtzW/zwHbO1olj7jnilxp\\ntXRNfBdA9iGNO3X2lWVuETbP4qOH4icwPzy5NKT9PnOLC2MJC9UNmUCewLWDdKo+\\nJZS857NR+QNiBG8dzWqKXXHXwknQt5aGdv8LbzdQZpoj2EvYONG9ckACXjFh4S9g\\n5EBoQbDm4Y38lIktZnx1tH92ofQ1aQ6pxQftF2wl/QBJjdoxiIbnmaTHfthF3sMe\\nW0+Rxkq3dC6szxjTGzLMzbnMBC2EUzF0+jh8Kmj2PstjMLeCt4X6azS6HcqTK0Md\\nSYnye+x6UDqHVb8BhJdbfMkOvuTdSvmUNOjDahCmSA65AQ0EXSoYJgEIAMoKpU1D\\nrda1GvHVwTjCfFF25kpadBhe6vVF2AEQu57GpqxJyKG2K38H+W3hv9sTcq7HXpVL\\nPLXVJ1hFmj9idaimG7xLu2GGlgoTYaA3R2of34sI1xS/37pwpAa9Qj6MG3xb1u7N\\ncsy4XZByNqT73uqz16iukyfhD0imE2q9x41batTFsTP5oHGp2BO5ki3F1xk2kcEb\\n4cI3jHq1sdqZzk48sS18Cur0QlbZKDPr/PmKgPzV9L/+ls2v3jA0TV7NLeXnhZ3V\\nJcdG3tZPXmT4Kvmr8jzrYlXul9T3DlKLU7IqYwvhUewXFkiowoilBsTy7ou0raDM\\naNkmcRiIkYgImQkAEQEAAYkBNgQYAQgAIBYhBPuaK2aExb0lImQAoIApP0QrhIIs\\nBQJdKhgmAhsMAAoJEIApP0QrhIIsMlwIAJP+C/YHaOa4usjzZk9X/ELm6e3Wjy2D\\nnQMxv84F0RgI8TAC0TniIRah6TvSJ2KgX9jCxcAhGdAtkWhgkSwjjHwT43oVXKlk\\ni3R3RJAzzR6h/q/lYFw5S3T8brA6POxz6KJQhiclocV+hCq8HxkEpfaoZubOpBok\\nQFnwDbhobs0f3Gvvr5Kto/rcyrhYuM1YkCayaTICQ1FUX9LAPFTyvJbUAViINkYU\\nkGcPHJc93NRiKemluWib6qp19K4DRPkT4RBw+P6OqmfiwAQ1T1fGXXrYc/pzkTqs\\nwgmSn4M9hTK7rt4UcZRKHFFHpaBay2BmksECSguf0r98cHTPiM/dTjE=\\n=3Iz2\\n-----END PGP PUBLIC KEY BLOCK-----"}
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /metrics                  --> github.com/zsais/go-gin-prometheus.prometheusHandler.func1 (4 handlers)
[GIN-debug] GET    /ping                     --> main.main.func1 (4 handlers)
[GIN-debug] GET    /                         --> main.main.func2 (4 handlers)
[GIN-debug] GET    /force-update             --> main.main.func3 (4 handlers)
[GIN-debug] Listening and serving HTTP on :4747
[GIN] 2019/07/13 - 17:45:51 | 200 |     268.434µs |      172.17.0.1 | GET      /ping
[GIN] 2019/07/13 - 17:46:27 | 200 |     162.523µs |      172.17.0.1 | GET      /
[GIN] 2019/07/13 - 17:47:17 | 200 |     205.894µs |      172.17.0.1 | GET      /
2019/07/13 17:47:17 &main.RandoRequest{Count:10, KvHosts:[]string{"http://etcd0.zts-poc.svc:2379/"}, KvKey:"/simple-client0/secrets.json", PubKey:"-----BEGIN PGP PUBLIC KEY BLOCK-----\\n\\nmQENBF0qGCYBCADLjTH3IqfcVs0X/x4xN+7AT+6cgvy7sYNHMCDbo2h/KCVRJo6c\\nmqw44gbTktzxzs3SugJFVIaWpsSsYO+VcmSvmufojnNmP5Sk4YVcM4MSXs0BouN1\\noCSw8gAzWh8rGEd5suVTzyX6YFUVO6RrjxVv9VkuFTmfotgQyEtxcO83YgqCZaqI\\nIlSej5Cfnqx4ccoH4fBISz+B07aINRZy3H5F27QGbG3g26hhcVo8GChy0hwl4psQ\\ns3AEJpWIhHFZpHfMXZkEWK1+JbbBS7aerDxJFDwA1HgnhhmWmuKEB70FhCl8bw0K\\n4OnSzlqS0kkfimUd5FqnNacaEEmzQk5+1JHzABEBAAG0LUNKIEhhcnJpZXMgKFpl\\ncm8gVHJ1c3QgU2VjcmV0cykgPGNqQHdvdHcucHJvPokBTgQTAQgAOBYhBPuaK2aE\\nxb0lImQAoIApP0QrhIIsBQJdKhgmAhsDBQsJCAcCBhUKCQgLAgQWAgMBAh4BAheA\\nAAoJEIApP0QrhIIsot8H/jtTV8ndJA83iB0Cf2oXnRFEtzW/zwHbO1olj7jnilxp\\ntXRNfBdA9iGNO3X2lWVuETbP4qOH4icwPzy5NKT9PnOLC2MJC9UNmUCewLWDdKo+\\nJZS857NR+QNiBG8dzWqKXXHXwknQt5aGdv8LbzdQZpoj2EvYONG9ckACXjFh4S9g\\n5EBoQbDm4Y38lIktZnx1tH92ofQ1aQ6pxQftF2wl/QBJjdoxiIbnmaTHfthF3sMe\\nW0+Rxkq3dC6szxjTGzLMzbnMBC2EUzF0+jh8Kmj2PstjMLeCt4X6azS6HcqTK0Md\\nSYnye+x6UDqHVb8BhJdbfMkOvuTdSvmUNOjDahCmSA65AQ0EXSoYJgEIAMoKpU1D\\nrda1GvHVwTjCfFF25kpadBhe6vVF2AEQu57GpqxJyKG2K38H+W3hv9sTcq7HXpVL\\nPLXVJ1hFmj9idaimG7xLu2GGlgoTYaA3R2of34sI1xS/37pwpAa9Qj6MG3xb1u7N\\ncsy4XZByNqT73uqz16iukyfhD0imE2q9x41batTFsTP5oHGp2BO5ki3F1xk2kcEb\\n4cI3jHq1sdqZzk48sS18Cur0QlbZKDPr/PmKgPzV9L/+ls2v3jA0TV7NLeXnhZ3V\\nJcdG3tZPXmT4Kvmr8jzrYlXul9T3DlKLU7IqYwvhUewXFkiowoilBsTy7ou0raDM\\naNkmcRiIkYgImQkAEQEAAYkBNgQYAQgAIBYhBPuaK2aExb0lImQAoIApP0QrhIIs\\nBQJdKhgmAhsMAAoJEIApP0QrhIIsMlwIAJP+C/YHaOa4usjzZk9X/ELm6e3Wjy2D\\nnQMxv84F0RgI8TAC0TniIRah6TvSJ2KgX9jCxcAhGdAtkWhgkSwjjHwT43oVXKlk\\ni3R3RJAzzR6h/q/lYFw5S3T8brA6POxz6KJQhiclocV+hCq8HxkEpfaoZubOpBok\\nQFnwDbhobs0f3Gvvr5Kto/rcyrhYuM1YkCayaTICQ1FUX9LAPFTyvJbUAViINkYU\\nkGcPHJc93NRiKemluWib6qp19K4DRPkT4RBw+P6OqmfiwAQ1T1fGXXrYc/pzkTqs\\nwgmSn4M9hTK7rt4UcZRKHFFHpaBay2BmksECSguf0r98cHTPiM/dTjE=\\n=3Iz2\\n-----END PGP PUBLIC KEY BLOCK-----"}
[GIN] 2019/07/13 - 17:47:17 | 200 |    15.49841ms |      172.17.0.1 | GET      /force-update
[GIN] 2019/07/13 - 17:47:47 | 200 |      95.556µs |      172.17.0.1 | GET      /
```
