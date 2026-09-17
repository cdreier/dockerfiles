# Alpine uses musl, not glibc. A default `go build` on this host links
# against /lib64/ld-linux-x86-64.so.2 and the container exits with
# "no such file or directory". CGO_ENABLED=0 produces a static binary.

BINARY ?= server
IMAGE  ?= drailing/singlehostreverseproxy
GOOS   ?= linux
GOARCH ?= amd64

.PHONY: build test docker clean

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags='-s -w' -o $(BINARY) .

test:
	go test ./...

docker: build
	docker build -t $(IMAGE) .

push: docker
	docker push $(IMAGE):latest

clean:
	rm -f $(BINARY)
