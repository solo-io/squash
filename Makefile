DOCKER_REPO ?= soloio
VERSION ?= $(shell git describe --tags)

.PHONY: all
all: binaries deployment

.PHONY: binaries
binaries: target/squash-client/squash-client target/squash


RELEASE_BINARIES := target/squash-client/squash-client target/squash-linux target/squash-osx target/squash-windows

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: manifests
manifests: deployment
	cp -f target/kubernetes/squash-client.yml ./contrib/kubernetes

.PHONY: upload-release
upload-release: release-binaries manifests dist
	./contrib/github-release.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=squash tag=$(VERSION)
	@$(foreach BINARY,$(RELEASE_BINARIES),./contrib/upload-github-release-asset.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=squash tag=$(VERSION) filename=$(BINARY);)

.PHONY: containers
containers: target/squash-client-container

.PHONY: prep-containers
prep-containers: target/squash-client/squash-client target/squash-client/Dockerfile


SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

target:
	[ -d $@ ] || mkdir -p $@

target/squash: target $(SRCS)
	go build -o $@ ./cmd/squash-cli

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

target/kubernetes/squash-client.yml: target/squash-client-container

target/kubernetes/:
	[ -d $@ ] || mkdir -p $@

deployment: | target/kubernetes/
deployment: target/kubernetes/squash-client.yml


.PHONY: clean
clean:
	rm -rf target

dist: target/squash-client-container
	docker push $(DOCKER_REPO)/squash-client:$(VERSION)

# make the solo-kit-provided resources
# do this on initialization and whenever the apichanges
.PHONY: generate-sk
generate-sk: docs-and-code/v1

docs-and-code/v1:
	go run cmd/generate-code/main.go

.PHONY: tmpclient
tmpclient:
	GOOS=linux go build -o target/squash-client/squash-client cmd/squash-client/platforms/kubernetes/main.go
