FROM golang:latest AS builder
MAINTAINER zengqiang96 "zengqiang96@gmail.com"

# 参考 https://yryz.net/post/golang-docker-alpine-start-panic.html
ENV CGO_ENABLED 0

WORKDIR /src
COPY . .

RUN \
    GOPROXY="https://goproxy.io" \
    go build .

FROM alpine:latest

COPY --from=builder /src/msnowflake /usr/bin/msnowflake

VOLUME ["/var/msnowflake"]

CMD ["-conf=/var/msnowflake/msnowflake.yaml"]
ENTRYPOINT ["/usr/bin/msnowflake"]

