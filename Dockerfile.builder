FROM golang:1.10
RUN apt-get update && apt-get -y install build-essential libsystemd-dev
WORKDIR /go/src/bsycorp/log-forwarder
