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
ENV         SECRETS_KEY '/simple-client/secrets.json'
ENV         SECRETS_COUNT 10
ENV         ETCD_HOSTS 'http://etcd0:2379/'

ENTRYPOINT ["/appuser/server"]
