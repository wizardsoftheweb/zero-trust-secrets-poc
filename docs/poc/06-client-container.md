# 06 Client Container

If the control server requires a GPG pub key to encrypt config, the clients need to be able to manage GPG keys. For this stage, I'd like to build a minimal container using an Ubuntu LTS image that clients can be dropped into. The container needs the following:

* `gpg{,2}` must be on the box
* The primary keyring must be built automatically
* The container should expose uid and gid for a new user to use the keyring so clients can be built with service users instead of running as `root`

## What I Actually Accomplished

From what I've gathered, it's a better pattern with Go containers to do [a multistage build](https://docs.docker.com/develop/develop-images/multistage-build/). That means reusing images requires more scripting than I feel like putting into the PoC. Needless to say, with the right `Makefile` and some judicious variablization, you could easily reuse your `Dockerfile` instead of making one for each binary.

For my container process, I (once again) adapted [this great project](https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324). Instead of starting with a specific hash, I just chose `golang:alpine3.10`. This is a PoC. It's vanilla aside from a home directory for `appuser` which will contain the keys. The second stage, however, starts from `alpine:3.10`. Something larger than `scratch` is needed to get `gnupg` in the final container. It's also useful to get in the box and do things on the running container before nuking it. Copying the directory loses its perms (unless I missed some flags?) so I had to `chown` it. Finally, setting the `WORKDIR` to `/appuser` means the client can start in `appuser` land and edit all the files it needs. The ports are client-specific and don't have to be ported around.

To get the client to talk to the control server, I had to snag my host ip using [this quick command](https://nickjanetakis.com/blog/docker-tip-65-get-your-docker-hosts-ip-address-from-in-a-container).
```shell-session
docker run \
    -p 127.0.0.1:4848:4747 \
    -e "RANDO_ENDPOINT=http://$(ip -4 addr show docker0 | grep -Po 'inet \K[\d.]+'):8080/rando" \
    -it \
    simple-client
```

Notice that...

1. The primary keyring is not built automatically: assuming clients are capable of bootstrapping themselves, they handle this. I discovered I was doubling up on the work. Personally I think it makes more sense for the application to set up its secrets than the application's environment.
2. There's no user customization: normally user granularity is needed when you're stacking images and processes (at least that where I normally use it). The image isn't really meant to be used as a base and it's easy enough to rebuild so I left that out entirely.

## The `Dockerfile`

I originally had this in its own directory and in simple client. That doesn't make sense. This is what I finished up with. What's live might be different.

```dockerfile
FROM        golang:alpine3.10 as builder

RUN         apk add --no-cache \
                ca-certificates \
                git \
                tzdata && \
            update-ca-certificates && \
            adduser -D -g '' --home '/appuser' appuser

WORKDIR     $GOPATH/src/wotw/client-container/
COPY        . .
RUN         go get -d -v
RUN         CGO_ENABLED=0 GOOS=linux go build \
                -ldflags="-w -s" \
                -a \
                -installsuffix cgo \
                -o /go/bin/server \
                .

FROM        alpine:3.10

COPY        --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY        --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY        --from=builder /etc/passwd /etc/passwd
COPY        --from=builder /appuser /appuser
COPY        --from=builder /go/bin/server /appuser/server

RUN         apk add --update \
                gnupg && \
            rm -rf /var/cache/apk/* && \
            chown -R appuser /appuser

USER        appuser
WORKDIR     /appuser
ENV         PORT=4747
EXPOSE      $PORT
ENV         RANDO_ENDPOINT ''

ENTRYPOINT ["/appuser/server"]
```
