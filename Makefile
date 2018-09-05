
IMG=log-forwarder-builder
SRCDIR=/go/src/bsycorp/log-forwarder

all: compile

image:
	docker build --file Dockerfile.builder -t $(IMG) .

clean:
	rm -rf build

compile: clean
	docker run --rm -v $(PWD):$(SRCDIR) -v /var/run/docker.sock:/var/run/docker.sock -w $(SRCDIR) $(IMG) sh -c 'export GOPATH=`pwd`; go get -d -t && go test -v && go build -v -o build/log-forwarder'

try:
	docker run -it --rm -v $(PWD):$(SRCDIR) -w $(SRCDIR) $(IMG) /bin/bash
