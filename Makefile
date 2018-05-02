.PHONY: all
all: binaries deployment

.PHONY: binaries
binaries: target/squash-server/squash-server target/squash-client/squash-client target/squash target/squash-lite-container/squash-lite-container target/squash-lite

.PHONY: release-binaries
release-binaries: target/squash-server/squash-server target/squash-client/squash-client target/squash-linux target/squash-osx target/squash-windows

.PHONY: containers
containers: target/squash-server-container target/squash-client-container

.PHONY: prep-containers
prep-containers: ./target/squash-server/squash-server target/squash-server/Dockerfile target/squash-client/squash-client target/squash-client/Dockerfile

DOCKER_REPO ?= soloio
VERSION ?= $(shell git describe --tags)


SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

target:
	[ -d $@ ] || mkdir -p $@

target/squash: target $(SRCS)
	go build -o $@ ./cmd/squash-cli

target/squash-lite: target $(SRCS)
	go build -ldflags "-X github.com/solo-io/squash/pkg/lite/kube.ImageVersion=$(VERSION) -X github.com/solo-io/squash/pkg/lite/kube.ImageRepo=$(DOCKER_REPO)" -o $@ ./cmd/squash-lite

target/squash-linux: target $(SRCS)
	GOOS=linux go build -o $@ ./cmd/squash-cli

target/squash-osx: target $(SRCS)
	GOOS=darwin go build -o $@ ./cmd/squash-cli

target/squash-windows: target $(SRCS)
	GOOS=windows go build -o $@ ./cmd/squash-cli

target/squash-client/: | target
target/squash-client/:
	[ -d $@ ] || mkdir -p $@

target/squash-client/squash-client: | target/squash-client/
target/squash-client/squash-client: $(SRCS)
	GOOS=linux CGO_ENABLED=0 go build -ldflags '-w' -o target/squash-client/squash-client ./cmd/squash-client/platforms/kubernetes

target/squash-client/Dockerfile: | target/squash-client/

target/squash-client/Dockerfile: ./cmd/squash-client/platforms/kubernetes/Dockerfile
	cp -f ./cmd/squash-client/platforms/kubernetes/Dockerfile ./target/squash-client/Dockerfile

target/squash-server/:
	[ -d $@ ] || mkdir -p $@

target/squash-server/squash-server: | target/squash-server/
target/squash-server/Dockerfile:    | target/squash-server/

target/squash-server/squash-server: $(SRCS)
	GOOS=linux CGO_ENABLED=0  go build -ldflags '-w' -o ./target/squash-server/squash-server ./cmd/squash-server/

target/squash-server/Dockerfile: cmd/squash-server/Dockerfile
	cp cmd/squash-server/Dockerfile target/squash-server/Dockerfile


target/squash-server-container: ./target/squash-server/squash-server target/squash-server/Dockerfile
	docker build -t $(DOCKER_REPO)/squash-server:$(VERSION) ./target/squash-server/
	touch $@

target/squash-lite-container/:
	[ -d $@ ] || mkdir -p $@

target/squash-lite-container/squash-lite-container: | target/squash-lite-container/
target/squash-lite-container/squash-lite-container: $(SRCS)
	GOOS=linux CGO_ENABLED=0  go build -ldflags '-w' -o ./target/squash-lite-container/squash-lite-container ./cmd/squash-lite-container/


target/squash-lite-container/Dockerfile.dlv:    | target/squash-lite-container/
target/squash-lite-container/Dockerfile.dlv: cmd/squash-lite-container/Dockerfile.dlv
	cp cmd/squash-lite-container/Dockerfile.dlv target/squash-lite-container/Dockerfile.dlv
target/squash-lite-container-dlv-container: ./target/squash-lite-container/squash-lite-container target/squash-lite-container/Dockerfile.dlv
	docker build -f target/squash-lite-container/Dockerfile.dlv -t $(DOCKER_REPO)/squash-lite-container-dlv:$(VERSION) ./target/squash-lite-container/
	touch $@
target/squash-lite-container-dlv-pushed: target/squash-lite-container-dlv-container
	docker push $(DOCKER_REPO)/squash-lite-container-dlv:$(VERSION)
	touch $@



target/squash-lite-container/Dockerfile.gdb:    | target/squash-lite-container/
target/squash-lite-container/Dockerfile.gdb: cmd/squash-lite-container/Dockerfile.gdb
	cp cmd/squash-lite-container/Dockerfile.gdb target/squash-lite-container/Dockerfile.gdb
target/squash-lite-container-gdb-container: ./target/squash-lite-container/squash-lite-container target/squash-lite-container/Dockerfile.gdb
	docker build -f target/squash-lite-container/Dockerfile.gdb -t $(DOCKER_REPO)/squash-lite-container-gdb:$(VERSION) ./target/squash-lite-container/
	touch $@
target/squash-lite-container-gdb-pushed: target/squash-lite-container-gdb-container
	docker push $(DOCKER_REPO)/squash-lite-container-gdb:$(VERSION)
	touch $@



target/squash-client-container: target/squash-client/squash-client target/squash-client/Dockerfile
	docker build -t $(DOCKER_REPO)/squash-client:$(VERSION) ./target/squash-client/
	touch $@

target/squash-client-base-container:
	docker build -t $(DOCKER_REPO)/squash-client-base -f cmd/squash-client/platforms/kubernetes/Dockerfile.base cmd/squash-client/platforms/kubernetes/
	touch $@

.PHONY: push-client-base
push-client-base:
	docker push $(DOCKER_REPO)/squash-client-base

target/%.yml : contrib/%.yml.tmpl
	SQUASH_REPO=$(DOCKER_REPO) SQUASH_VERSION=$(VERSION) go run contrib/templategen.go $< > $@

target/kubernetes/squash-server.yml: target/squash-server-container
target/kubernetes/squash-client.yml: target/squash-client-container

target/kubernetes/:
	[ -d $@ ] || mkdir -p $@

deployment: | target/kubernetes/
deployment: target/kubernetes/squash-client.yml target/kubernetes/squash-server.yml


.PHONY: clean
clean:
	rm -rf target

pkg/restapi: api.yaml
	swagger generate server --name=Squash --exclude-main --target=./pkg/  --spec=./api.yaml
	swagger generate client --name=Squash --target=./pkg/  --spec=./api.yaml

dist: target/squash-server-container target/squash-client-container
	docker push $(DOCKER_REPO)/squash-client:$(VERSION)
	docker push $(DOCKER_REPO)/squash-server:$(VERSION)