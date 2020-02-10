FROM golang:latest AS builder
MAINTAINER zengqiang96 "zengqiang96@gmail.com"

# 参考 https://yryz.net/post/golang-docker-alpine-start-panic.html
ENV CGO_ENABLED 0

WORKDIR /src

RUN  \
     GOPROXY="https://goproxy.io" && \
     apk add --no-cache git && \
     git clone https://github.com/Lazzy/msnowflake && cd msnowflake && \
     go build .

FROM alpine:latest

COPY --from=builder /src/msnowflake/msnowflake /usr/bin/msnowflake

VOLUME ["/var/msnowflake"]

CMD ["-conf=/var/msnowflake/msnowflake.yaml"]
ENTRYPOINT ["/usr/bin/msnowflake"]

