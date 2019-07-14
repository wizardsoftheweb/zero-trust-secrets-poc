# 09 The First Full Demo

I wanted to pause and review the goals I laid out in [the root `README`](./../../README.md#poc-onboarding). I've duplicated them here, because if I've met them, it's time to do more cool stuff.

## The Goals

I wanted to build something that...

1) provides an automated and centralized secret generation service,
2) provides different secrets to each environment while still remaining in the same pipeline,
3) provides a mechanism to update secrets in all environments automatically, and
4) provides a centralized location to monitor the status (not the contents) of the secrets.

###  Automated and Centralized Secrets

This is currently met in two parts. The control server is the primary hub of this PoC. With some development time, it would be very simple to add these things:

1) server authentication: only accept requests from trusted sources
2) config signatures: this should be easy to do but will require another library or two
3) myriad secrets: I feel like it's pretty obvious that changing things up won't be hard
4) scheduled updates: `cron` those jobs

The control server is not automated, though. Through the course of this PoC I've delegated most of that automation to the clients needing secrets. It makes more sense to me that the control server serves as a trusted generation point but doesn't know what it's generating. So far automating client secret generation is very easy.

### Different Secrets in Different Places

This is handily done. The secrets are so different there's no way to reproduce or replicate any of them. Every new thing that stands up can get its own unique secrets not seen or touched by anything else. As a huge proponent of good crypto and privacy on the web, I am stoked. As someone who might have to debug this later, I'm willing to admit there are a few things that should be touched up.

1) Audits: The initial keys should be passed off to another trusted, central location for reviewing config when things go wrong.
2) No replication: Secrets are unique to a single instance of any client. That means you can't replicate instances to increase throughput. If you're totally stateless, that's not a huge deal but there's a lot of other tooling that needs to be built to handle that. If you aren't stateless, you need replication at some point.
3) `gpg` is a huge pain in the ass: It was fast, simple, and got the MVP out the door. Dunno that I like it longterm.

### Automatic Updates

The resolution here is very similar to the first goal's resolution. The control server can reach out and update anything it's asked to with whatever values. A lot of that logic has been delegated out to the clients, at least for now. Using Viper, they can easily periodically update themselves. They'll also know if the remote config changes and they should adjust themselves automatically.

There is something to be said for some sort of secret orchestration solution on top of the control and clients. I think a necessary first step is figuring out how to share files such as the keys in such a way that the storage doesn't backdoor the encrypted config. It also requires some thought about what gets stored and how. If everything is totally ephemeral and unique, tasks such as scheduling key rotation become impossible because there's no way to track and organize them.

### Central Monitoring

This is the handwaviest resolution. I've got Prometheus metrics on everything. For the PoC, I didn't do much logging. I certainly didn't set things up well to automated recovery. To me, those things are part of actually committing to a project instead of trying to slap together something that works. I'm not deploying this anywhere yet so I've brushed those things off.

To some extent, this is also stuck in the same limbo as the previous two. To monitor status you need to track everything. You also need to store things. Figuring out the storage question and defining a good way to track all the moving pieces will give this goal a better foundation.

## The Demo

### Prereqs

You'll want [`kubectx`/`kubens`](https://github.com/ahmetb/kubectx).

These are assumed to be run before anything else.
```shell-session
minikube start --profile pocDemo --cpus 4 --memory 4096 
kubectx pocDemo
alias k=kubectl
```

### Checking out the Deployment

```shell-session
$ cd path/to/zero-trust-secrets

$ k get ns
NAME              STATUS   AGE
default           Active   2m32s
kube-node-lease   Active   2m35s
kube-public       Active   2m35s
kube-system       Active   2m35s

$ helm init --history-max 200
$HELM_HOME has been configured at /home/cjharries/.helm.

Tiller (the Helm server-side component) has been installed into your Kubernetes Cluster.

Please note: by default, Tiller is deployed with an insecure 'allow unauthenticated users' policy.
To prevent this, run `helm init` with the --tiller-tls-verify flag.
For more information on securing your installation see: https://docs.helm.sh/using_helm/#securing-your-helm-installation

$ helm install helm-charts/zts-poc
NAME:   nordic-bumblebee
LAST DEPLOYED: Sat Jul 13 21:49:43 2019
NAMESPACE: default
STATUS: DEPLOYED

RESOURCES:
==> v1/Deployment
NAME            READY  UP-TO-DATE  AVAILABLE  AGE
control-server  0/1    1           0          0s
etcd0           0/1    1           0          0s
etcd1           0/1    1           0          0s
simple-client0  0/1    1           0          0s

==> v1/Namespace
NAME     STATUS  AGE
zts-poc  Active  0s

==> v1/Pod(related)
NAME                             READY  STATUS             RESTARTS  AGE
control-server-6f8595dccf-xjnfh  0/1    Init:0/1           0         0s
etcd0-556999fdd8-rc8k9           0/1    ContainerCreating  0         0s
etcd1-67f55857bf-pr57t           0/1    ContainerCreating  0         0s
simple-client0-5d5bbb6bb9-nx4q6  0/1    Init:0/1           0         0s

==> v1/Service
NAME            TYPE      CLUSTER-IP      EXTERNAL-IP  PORT(S)                        AGE
control-server  NodePort  10.104.251.135  <none>       8080:32613/TCP                 0s
etcd-client     NodePort  10.99.67.209    <none>       2379:31189/TCP                 0s
etcd0           NodePort  10.108.116.39   <none>       2379:32550/TCP,2380:30178/TCP  0s
etcd1           NodePort  10.101.90.66    <none>       2379:31637/TCP,2380:31149/TCP  0s
simple-client0  NodePort  10.101.106.209  <none>       4747:30602/TCP                 0s


NOTES:
Thank you for installing zts-poc.

Your release is named nordic-bumblebee.

To learn more about the release, try:

  $ helm status nordic-bumblebee

$ kubens zts-poc
Context "pocDemo" modified.
Active namespace is "zts-poc".

$ k get po
control-server-6f8595dccf-xjnfh   1/1     Running   0          25s
etcd0-556999fdd8-rc8k9            1/1     Running   0          25s
etcd1-67f55857bf-pr57t            1/1     Running   0          25s
simple-client0-5d5bbb6bb9-nx4q6   1/1     Running   0          25s

$ k get svc
NAME             TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)                         AGE
control-server   NodePort   10.104.251.135   <none>        8080:32613/TCP                  34s
etcd-client      NodePort   10.99.67.209     <none>        2379:31189/TCP                  34s
etcd0            NodePort   10.108.116.39    <none>        2379:32550/TCP,2380:30178/TCP   34s
etcd1            NodePort   10.101.90.66     <none>        2379:31637/TCP,2380:31149/TCP   34s
simple-client0   NodePort   10.101.106.209   <none>        4747:30602/TCP                  34s

$ minikube service list --namespace zts-poc
|-----------|----------------|--------------------------------|
| NAMESPACE |      NAME      |              URL               |
|-----------|----------------|--------------------------------|
| zts-poc   | control-server | http://192.168.99.104:32613    |
| zts-poc   | etcd-client    | http://192.168.99.104:31189    |
| zts-poc   | etcd0          | http://192.168.99.104:32550    |
|           |                | http://192.168.99.104:30178    |
| zts-poc   | etcd1          | http://192.168.99.104:31637    |
|           |                | http://192.168.99.104:31149    |
| zts-poc   | simple-client0 | http://192.168.99.104:30602    |
|-----------|----------------|--------------------------------|

```

