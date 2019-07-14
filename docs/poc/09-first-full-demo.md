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

### Verifying Secrets

```shell-session
$ export CONTROL='http://192.168.99.104:32613'
$ export CLIENT='http://192.168.99.104:30602'
$ alias etcdemo='etcdctl --endpoints http://192.168.99.104:31189 --no-sync'
$ etcdemo ls /
/simple-client0

$ etcdemo get /simple-client0/secrets.json
wcBMA6/UdzJ72KTGAQgADtwbjy/QqX711NsFBw34fszY5yQ+drX4PjIiTueJzSuJSzzz3888HhDjKSPu/wTNSaFdsUyIkLnSlJi64UQmOu5L4OUBibVqUu2p1x+6dCh1sWY3jsDih9Jysn3S56IXX1enxRt5LtAq6L5fSufcOa0yMz7zbAP0k6stHGOjaNgI2bFKhss8Xg2eh69xVbjzi/Dcyv+8arJrSfp6hYZPdRGrttiLr2xFsjLbCvUmuLxVlLELHqVUrVpQoy94wZSo2ipDvgURqC63RNPdLl+y3rSvVPswWfv+JhEn/8t21mGmP5SW6u+x4IzfB/eBd62/622kzPk4UZf2JWaI1kVyEdLgAeSgG29nOEZhuveKvqj/fEzO4fX/4Fbg1eGsHOBX4onKg8fgheM/TMk+6QRDeeAL4Yt14P/n1oN2eq22AzI6xTuxnOYkNahZzCSP9+4ROai/EWZXeGic2Q3JgoDGxW6KkyHyRO2uOQ12AQ1SiSUGtf2MuleQWxHHf5dXl8chxckvtd+ULQukmP6SjvdZvugYhfYfUZTvjoAlKhVe3yTyJVtiAvmt0fP3AcVrG4IuACUzhE13Vibgx+ahAPIZKmoK/UPawSy+eNnyoj/ZhBRP3gWRKUnyfsIqM3qWLLTE08W+kY7R9wiQdaBjgtGzFMKn0TKeH1Hvqjj84IblVa+fcfVClIjanMAyvZNfpsR34smGDGrVrP9UALpmiIfgb+TgQOd1WsA+k1DWXORLZjTi4IznJs3Z6UJB6JXKiRQjOVNnlz52aeyx/cO620hi2t0j+ElTMpdIDkVoEFeZz4OGJ/mZlU14A9gW9vPUHvfXKEg2R1sjSsPgWCVLTNeH+Ik81UDXuAkswiMUbQCZRrtIpFSz16QdgO9t51yuzBbsWpUIqKZnIniBRzyrOtqSm0T83UzgquZYDT5fYoNZq5vWxE44w1rXqpX+BZppuJdQARiIrI+HIBPxVKY0FGlp5U8IBHwG1pLayRRt7WufQvEirwvPBXI84EDlH9W1n++IdYhT5xk2mu+j9iPoGzlPf/HGPeOh3OaR3DvgzeTYMii1R7DsPdgqSdpImzwB4OrmdetishcZ3Z8SlweYXvxyYWfV6K7MIj7K8zfFtDpR8BnVJOcyPpu8JIXjmTBTX9hfw89vgCGb7j2fQ1g+CHqzJOAm4lQDWg/gPeKUj06S4FPjxbur4tSZDJjgx+SLwEiC+3PPNdcx+OWEQzw64ngqLP3hWP0A

$ curl -s $CLIENT | jq
{
  "secrets": [
    "MR4ltlrAC5Bmr5AFaCfSoAoXMxzvDlayd3zaxsAhYMrw84J9LkZ5izu6shjxjeE=",
    "00WeYTsnd-Np4wlqV38kX0p_Oi7-386QYcKK7b22n85fc3t9uatzq84tzE2kIkI=",
    "PJMRmgjW50EJVyD-n4T0RtFqw2trhfupd6lbBdfT2Qxu5JoLTRS3FsLDsaS-b-w=",
    "pxXUV_F1XkQzUeVTcuER6Z_NGbA9npeQ9unHzLeMTSralMAE2PKtp8pbtMIyujQ=",
    "e9Gg28JEtfefrhZIdCq2PqaI_jbQbOoVByRFjTNj20740YMqX_pYPzWm2lb_NFM=",
    "rT0UUTHcy7qP-cJpm6bAJDszaLpNW8kEnbIQRebT_CVmNeXSveN6q3X796HY6Nk=",
    "rf0picJsrc5RdKjxYBVJ2BeLTJ7_ZRTi-BROhnpN0sqjRUbJD4ul6p4dI3D1sdk=",
    "rvEH91dSTwsG-wJDQRJem7KO04lPb8WbcevDCgYCwMtL9yt5-2PosDu4K_nvaTY=",
    "gZ2eRtEf-413W_lfPTS1RIzw5ZYpuVYxIeN4H8fryM-Tzjh7mtpgYWdOnZqUyYY=",
    "CKilTvIKwHtAzaW5K2f99t2B91f6WqV7WvzjUw1YMn5_P0rX1uO-iN4-Od3yKKY="
  ]
}

$ curl -S $CLIENT/force-update | jq
{
  "message": "Secrets were regenerated"
}

$ etcdemo get /simple-client0/secrets.json
wcBMA6/UdzJ72KTGAQgAPRWGDcJNJY8yPZKJBwOwm8XI8hOt5tCkQf6KphrMzfldkdv3/sz7nko9jnvBVS3XJOeIF5iWeXSV4bLaQi9UrctKh9sy9nfWxfCQGiSWQOJkIE4Qk3yhnygaEuwT6WUXi6dq07I6isl8kZZ5IDRKYcIa37y+Nnrhw0K3hYxu9NQw50BOfshCA/Hln8YwbsW1b5dDRL3BM/V6RjI7SeJ3mutP9BRUREPdgEI8CqxwphS3BbWgeEc7SB2JH7LDBnpbMGqAcQZeAhs9AUjvPV56zH+iEsA/IR4C7JjbLLIo5+fWMIlvtUTWs8Ua5wSAINJSiYhhPI4UqwrGKgkOUpchMtLgAeStXOkrj/RiEvfyZkTCO36b4Uva4AXgROGucOD44tGOk1DgGONAze25Z8deUeAv4etU4DDn4aK4BZMVmwp+GHRojxPnck8lEmBEShfaqR/P7O1WoyK5auno8ZO0y2BMU8b0NFhfzv8W+UT7ex94GSDfKOGdRmASkKF526ZTAXjFOOjjOgiExuYt2qjMZWhM89uABM+PJVAmS7BZjScPhO9A/L0U4nrNrHM0Jfkr7me4z4efwt/gW+bJl2DZRkkwkDRBjnydRysQehndl8zD8X6tuQ0DZBX8cMcmVCS2WhD1aw+k9LE2dpCqiIR8liIp+WA7Wtz5kY194DPlDVTZqFL3LczKjMGRO0P40eCGh30lEwdmmXbabWvNBEHgP+QC8q3RV2Kx+UIAY+FS0LiU4D3nvGq+xP+njvHoDjvh8ANO60d1zoHsEKAA8T5ZOg0CjDKI6SOk9q0nEtq7PISxuX9+Bs9bJrdGxd88/l93knpRcfS87rNp0VL5W8jhbwwrpnyPgOZlB1VSLBtVqIATV8ke4DNFX0VtlgLJCzVdZZ1ZS8n5/5wXUqp8RuuwpdKY0GzgFOYtKTYurxHKw6c/STZzrUS+3elPPnHQj2cDt1MHBJHJy1Uc8ZChjw1LQ94gKd43qxSMQHzZ1ipUayXM2EsXCfM+4DTlwGP2798+wH8dPhgw8m2vRfoPZ0wukOpemsPDs3ZnpADgluTSWXtlNdCFGM4GuG07pORF4CrmhaQdoCMnw9UB/mGcHmamzSXkYOxB75o+zkHjKNndLaTx8yyw7gnove9h3X77MD4cI0Wm9cxA2g0pv0Jw/ND6IuCA4sn3jQzgFuGGJuAM4qpcHk/gMeMY8kntgetAQ+Ax5F+njC85SdPS2uEYEK2mxffi8cUb/eFNUQA=

$ curl -s $CLIENT | jq
{
  "secrets": [
    "hh4zRDZzCZ4clYvLvXmh5ZEnj4iLxo3GSdT9AoCKBh0sabKWP3SCd024ztG0jCA=",
    "TiihFBbHUTmi1lWp8VHmmcRN30-N272ard_evjeiuqDZLWTqEQoC8-izlLP__p4=",
    "gIhLD2SVeCTCBU1b187MAmnKwOXxTB0E4IC5Hc6aUgKLx4tvf9bOK0PS-YnBQgo=",
    "PNzZKs_nA5kD67pJSt1xwvl-SA4MkePnmlHEjNcvExE75cGZhwBY1uvpAm-Mr7I=",
    "uEEGorV9e1d2lemV-1y34J_SPwuH-48hupnROiy5if-E1J5kQJ8tAbAsA9pnfjk=",
    "sk-Fp8wImpu47C2o5v1YPDbV8sZSKAtCKPM1J-4JfQ6fLMdJaa8-36LHZ7DH7Is=",
    "n4U6fJH4C1yykA7rczeZLUUUih6sC4n207sZLyt0a_KmeDpTNEaieEA1ZSOEt64=",
    "JiwOavdsjT7Ba0ACT0aeNrHAxhr39AGLlmKDvLymFtEdWFSHokKj8l3dmLFMPZw=",
    "7xc41Mf1yjGzWBIORdOf_tOW19Es-2lu14ap4lRQpsSfgTifsfAecEgRDgrt8AA=",
    "UVhGiz0ReGTIVHEBtJeUrnoOmCw7rS1NcAHDQxbcFA7rPk0yw6j14ZVr4SKvMm0="
  ]
}
```

### Verifying Metrics

I wanted a simple way to check this stuff out without having to run the full Promotheus operator on my cluster. So I snagged [`prom2json`](https://github.com/prometheus/prom2json) and skimmed the output.

```shell-session
$ prom2json $CONTROL/metrics \
    | jq -r '.[]|select(.name=="gin_requests_total")|[.metrics[]| {url: .labels.url, count: .value}]|group_by(.url)|[.[]|{url: .[0].url, count: map(.count | tonumber) | add}]|sort_by(.url)|.[]|.url + " " + (.count | tostring)' \
    | termgraph --title 'Requests Per Endpoint' --width 10 --custom-tick '🍸'

# Requests Per Endpoint

/           : 🍸 1.00 
/favicon.ico: 🍸 1.00 
/ping       : 🍸🍸 2.00 
/rando      : 🍸🍸🍸 3.00 

$ prom2json $CLIENT/metrics \
    | jq -r '.[]|select(.name=="gin_requests_total")|[.metrics[]| {url: .labels.url, count: .value}]|group_by(.url)|[.[]|{url: .[0].url, count: map(.count | tonumber) | add}]|sort_by(.url)|.[]|.url + " " + (.count | tostring)' \
    | termgraph --title 'Requests Per Endpoint' --width 10 --custom-tick '🍸'

# Requests Per Endpoint

/            : 🍸🍸🍸🍸 4.00 
/favicon.ico : 🍸 1.00 
/force-update: 🍸 1.00 
```

I used [`termgraph`](https://github.com/mkaz/termgraph) for the pretty pictures
