FROM debian:buster

MAINTAINER Tombo Kombo <tombo.hulata@gmail.com>

ENV GOPATH=/go
ENV LISTEN=0.0.0.0:6666
RUN \
    apt-get update && apt-get install -y golang \
    go-dep git && \
    mkdir -p /go/src/project;
ADD . /go/src/project/
WORKDIR /go/src/project/
RUN \
    dep ensure; \
    go build -o ./redis-sentinel-proxy; \
    cp redis-sentinel-proxy /usr/bin/redis-sentinel-proxy; \
    chmod +x /usr/bin/redis-sentinel-proxy;

EXPOSE 6666
CMD ["redis-sentinel-proxy"]
