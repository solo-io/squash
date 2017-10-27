.PHONY: all
all: target/squash-server/squash-server target/squash-client/squash-client target/squash deployment

.PHONY: containers
containers: target/squash-server-container target/squash-client-container

.PHONY: prep-containers
prep-containers: ./target/squash-server/squash-server target/squash-server/Dockerfile target/squash-client/squash-client target/squash-client/Dockerfile

DOCKER_REPO ?= soloio
VERSION ?= v0.1.2


SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

target:
	[ -d $@ ] || mkdir -p $@

target/squash: target $(SRCS)
	go build -o ./target/squash ./cmd/squash-cli

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

target/squash-client-container: target/squash-client/squash-client target/squash-client/Dockerfile
	docker build -t $(DOCKER_REPO)/squash-client:$(VERSION) ./target/squash-client/
	touch $@

target/%.yml : contrib/%.yml.tmpl
	SQUASH_REPO=$(DOCKER_REPO) SQUASH_VERSION=$(VERSION) go run contrib/templategen.go $< > $@

target/kubernetes/squash-server.yml: target/squash-server-container
target/kubernetes/squash-ds.yml: target/squash-client-container

target/kubernetes/:
	[ -d $@ ] || mkdir -p $@

deployment: | target/kubernetes/
deployment: target/kubernetes/squash-ds.yml target/kubernetes/squash-server.yml


.PHONY: clean
clean:
	rm -rf target

pkg/restapi: api.yaml
	swagger generate server --name=Squash --exclude-main --target=./pkg/  --spec=./api.yaml
	swagger generate client --name=Squash --target=./pkg/  --spec=./api.yaml

dist: target/squash-server-container target/squash-client-container
	docker push $(DOCKER_REPO)/squash-client:$(VERSION)
	docker push $(DOCKER_REPO)/squash-server:$(VERSION)