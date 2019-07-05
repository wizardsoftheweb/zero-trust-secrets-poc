# 01 Setting Up etcd

## PoC MVP

It's all local. Not even gonna mess with a local cluster. I figured that would be the fastest way.

### Installing

I specifically grabbed [v3.3.13](https://github.com/etcd-io/etcd/releases/tag/v3.3.13).

This is from the release page.
```bash
ETCD_VER=v3.3.13
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GITHUB_URL}

rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
rm -rf /tmp/etcd-download-test && mkdir -p /tmp/etcd-download-test

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /tmp/etcd-download-test --strip-components=1
rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

/tmp/etcd-download-test/etcd --version
ETCDCTL_API=3 /tmp/etcd-download-test/etcdctl version
```
I updated it to use `/srv/etcd` instead. I like [the `/srv` directory](http://refspecs.linuxfoundation.org/FHS_3.0/fhs/ch03s17.html). Also [`PrivateTmp`](https://www.freedesktop.org/software/systemd/man/systemd.exec.html#PrivateTmp=) is a thing.
```bash
sudo su -
mkdir -p /srv/etcd/lib

ETCD_VER=v3.3.13
GITHUB_URL=https://github.com/etcd-io/etcd/releases/download
DOWNLOAD_URL=${GITHUB_URL}

rm -f /srv/etcd/etcd-${ETCD_VER}-linux-amd64.tar.gz
rm -rf /srv/etcd/etcd-download-test && mkdir -p /srv/etcd

curl -L ${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz -o /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz
tar xzvf /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz -C /srv/etcd --strip-components=1
rm -f /tmp/etcd-${ETCD_VER}-linux-amd64.tar.gz

/srv/etcd/etcd --version
ETCDCTL_API=3 /srv/etcd/etcdctl version

u_g_id=$((16#$(openssl rand -hex 2)))
(getent group $u_g_id || getent passwd $u_g_id) || groupadd -g $u_g_id etcd && useradd -u $u_g_id -g etcd -d /srv/etcd -s /bin/false etcd
chown -R etcd:etcd /srv/etcd 
```

This didn't seem to install any sort of unit file so after messing with building my own from Stack Overflow and fighting SELinux on everything I just went with this solution.

```bash
dnf install -y etcd
```

I'm not a smart man. I'm not used to things that actually come in a package manager anymore.

```
: enabled capabilities for version 3.2
: published {Name:default ClientURLs:[http://localhost:2379]} to cluster cdf818194e3a8c32
: ready to serve client requests
: serving insecure client requests on 127.0.0.1:2379, this is strongly discouraged!
: Started Etcd Server.
```

That's perfect!

### Using It

I'm not going to spend any time on this.

```shell-session
$ etcdctl ls

$ etcdctl mk /my/first/key my-first-value
my-first-value
$ etcdctl ls
my
$ etcdctl ls --recursive
/my
/my/first
/my/first/key
```
