# Check out the project this is from
# https://github.com/chemidy/smallest-secured-golang-docker-image/blob/master/Dockerfile
# It's really neat
#
# I'm well aware I'm completely mitigated the benefits of the project by pulling it out
# I do what I want
FROM golang@sha256:cee6f4b901543e8e3f20da3a4f7caac6ea643fd5a46201c3c2387183a332d989 as builder
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates
RUN adduser -D -g '' appuser
WORKDIR $GOPATH/src/wotw/control-server/
COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/server .

FROM alpine:3.10
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /go/bin/server /go/bin/server

RUN chown -R appuser /tmp

USER appuser

ENV PORT=8080
EXPOSE $PORT

ENTRYPOINT ["/go/bin/server"]
