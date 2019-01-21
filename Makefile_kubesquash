.PHONY: all
all: binaries containers

DOCKER_REPO ?= soloio
VERSION ?= $(shell git describe --tags)

.PHONY: binaries
binaries: target/kubesquash-container/kubesquash-container target/kubesquash


RELEASE_BINARIES := target/kubesquash-linux target/kubesquash-osx 

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: containers
containers: target/kubesquash-container-dlv-container target/kubesquash-container-gdb-container 

.PHONY: push-containers
push-containers: target/kubesquash-container-dlv-pushed target/kubesquash-container-gdb-pushed

.PHONY: release
release: push-containers release-binaries

.PHONY: upload-release
upload-release: release
	./hack/github-release.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=kubesquash tag=$(VERSION)
	@$(foreach BINARY,$(RELEASE_BINARIES),./hack/upload-github-release-asset.sh github_api_token=$(GITHUB_TOKEN) owner=solo-io repo=kubesquash tag=$(VERSION) filename=$(BINARY);)

SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

target:
	[ -d $@ ] || mkdir -p $@

target/kubesquash: target $(SRCS)
	go build -ldflags "-X github.com/solo-io/kubesquash/pkg/cmd.ImageVersion=$(VERSION) -X github.com/solo-io/kubesquash/pkg/cmd.ImageRepo=$(DOCKER_REPO)" -o $@ ./cmd/kubesquash

target/kubesquash-osx: target $(SRCS)
	GOOS=darwin go build -ldflags "-X github.com/solo-io/kubesquash/pkg/cmd.ImageVersion=$(VERSION) -X github.com/solo-io/kubesquash/pkg/cmd.ImageRepo=$(DOCKER_REPO)" -o $@ ./cmd/kubesquash

target/kubesquash-linux: target $(SRCS)
	GOOS=linux go build -ldflags "-X github.com/solo-io/kubesquash/pkg/cmd.ImageVersion=$(VERSION) -X github.com/solo-io/kubesquash/pkg/cmd.ImageRepo=$(DOCKER_REPO)" -o $@ ./cmd/kubesquash

target/kubesquash-container/:
	[ -d $@ ] || mkdir -p $@

target/kubesquash-container/kubesquash-container: | target/kubesquash-container/
target/kubesquash-container/kubesquash-container: $(SRCS)
	GOOS=linux CGO_ENABLED=0  go build -ldflags '-w' -o ./target/kubesquash-container/kubesquash-container ./cmd/kubesquash-container/


target/kubesquash-container/Dockerfile.dlv:    | target/kubesquash-container/
target/kubesquash-container/Dockerfile.dlv: cmd/kubesquash-container/Dockerfile.dlv
	cp cmd/kubesquash-container/Dockerfile.dlv target/kubesquash-container/Dockerfile.dlv
target/kubesquash-container-dlv-container: ./target/kubesquash-container/kubesquash-container target/kubesquash-container/Dockerfile.dlv
	docker build -f target/kubesquash-container/Dockerfile.dlv -t $(DOCKER_REPO)/kubesquash-container-dlv:$(VERSION) ./target/kubesquash-container/
	touch $@

target/kubesquash-container-dlv-pushed: target/kubesquash-container-dlv-container
	docker push $(DOCKER_REPO)/kubesquash-container-dlv:$(VERSION)
	touch $@



target/kubesquash-container/Dockerfile.gdb:    | target/kubesquash-container/
target/kubesquash-container/Dockerfile.gdb: cmd/kubesquash-container/Dockerfile.gdb
	cp cmd/kubesquash-container/Dockerfile.gdb target/kubesquash-container/Dockerfile.gdb
target/kubesquash-container-gdb-container: ./target/kubesquash-container/kubesquash-container target/kubesquash-container/Dockerfile.gdb
	docker build -f target/kubesquash-container/Dockerfile.gdb -t $(DOCKER_REPO)/kubesquash-container-gdb:$(VERSION) ./target/kubesquash-container/
	touch $@
target/kubesquash-container-gdb-pushed: target/kubesquash-container-gdb-container
	docker push $(DOCKER_REPO)/kubesquash-container-gdb:$(VERSION)
	touch $@



.PHONY: clean
clean:
	rm -rf target

dist: target/kubesquash-container-gdb-pushed target/kubesquash-container-dlv-pushed