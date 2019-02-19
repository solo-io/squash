.PHONY: all
all: binaries containers ## (default) Builds binaries and containers

DOCKER_REPO ?= soloio
VERSION ?= $(shell git describe --tags)

.PHONY: help
help:
	 @echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sort | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"

.PHONY: binaries
binaries: target/kubesquash-container/kubesquash-container target/kubesquash # Builds kubesquash binaries in and places them in target/ folder

RELEASE_BINARIES := target/kubesquash-linux target/kubesquash-osx

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: containers
containers: target/kubesquash-container-dlv-container target/kubesquash-container-gdb-container ## Builds kubesquash debug containers

.PHONY: push-containers
push-containers: target/kubesquash-container-dlv-pushed target/kubesquash-container-gdb-pushed ## Pushes kubesquash debug containers to $(DOCKER_REPO)

.PHONY: release
release: push-containers release-binaries ## Pushes containers to $(DOCKER_REPO) and releases binaries to GitHub

.PHONY: upload-release
upload-release: ## Uploads artifacts to GitHub releases
	./hack/github-release.sh owner=solo-io repo=kubesquash tag=$(VERSION)
	@$(foreach BINARY,$(RELEASE_BINARIES),./hack/upload-github-release-asset.sh owner=solo-io repo=kubesquash tag=$(VERSION) filename=$(BINARY);)

SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

# Pass in build-time variables
LDFLAGS=-ldflags "-X main.KubeSquashVersion=$(VERSION) -X github.com/solo-io/kubesquash/pkg/cmd.ImageVersion=$(VERSION) -X github.com/solo-io/kubesquash/pkg/cmd.ImageRepo=$(DOCKER_REPO)"

target:
	[ -d $@ ] || mkdir -p $@

target/kubesquash: target $(SRCS)
	go build ${LDFLAGS} -o $@ ./cmd/kubesquash

target/kubesquash-osx: target $(SRCS)
	GOOS=darwin go build ${LDFLAGS} -o $@ ./cmd/kubesquash

target/kubesquash-linux: target $(SRCS)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo ${LDFLAGS} -o $@ ./cmd/kubesquash

target/kubesquash-container/:
	[ -d $@ ] || mkdir -p $@

target/kubesquash-container/kubesquash-container: | target/kubesquash-container/
target/kubesquash-container/kubesquash-container: $(SRCS)
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o ./target/kubesquash-container/kubesquash-container ./cmd/kubesquash-container/


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

.PHONY: publish-extension
publish-extension: bump-extension-version ## (vscode) Publishes extension
	./hack/publish-extension.sh

.PHONY: package-extension
package-extension: bump-extension-version ## (vscode) Packages extension
	cd extension/vscode && vsce package

.PHONY: bump-extension-version
bump-extension-version:  ## (vscode) Bumps extension version
	cd extension/vscode && \
	jq '.version="$(VERSION)" | .version=.version[1:]' package.json > package.json.tmp && \
	mv package.json.tmp package.json && \
	jq '.version="$(VERSION)" | .binaries.linux="$(shell sha256sum target/kubesquash-linux|cut -f1 -d" ")" | .binaries.darwin="$(shell sha256sum target/kubesquash-osx|cut -f1 -d" ")"' src/squash.json > src/squash.json.tmp && \
	mv src/squash.json.tmp src/squash.json

.PHONY: clean
clean: ## Deletes target folder
	rm -rf target

dist: target/kubesquash-container-gdb-pushed target/kubesquash-container-dlv-pushed ## Pushes all containers to $(DOCKER_REPO)