# Build stage
FROM golang:1.12 AS build-env
RUN apt-get update && apt-get -y install build-essential libsystemd-dev
WORKDIR /app/log-forwarder

# Populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go test -v
RUN go build -v -o build/log-forwarder

# Run stage
FROM bitnami/minideb:stretch
# doesnt use pinned version as we rely on content trust
ENV LANG C.UTF-8
RUN mkdir -p /var/lib/log-forwarder
COPY --from=build-env /app/log-forwarder/build/log-forwarder /opt/bin/log-forwarder
RUN chmod +x /opt/bin/log-forwarder && chmod -R 666 /var/lib/log-forwarder
WORKDIR /var/lib/log-forwarder
CMD /opt/bin/log-forwarder
