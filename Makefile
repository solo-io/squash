DOCKER_REPO ?= soloio
# VERSION ?= $(shell git describe --tags)
VERSION = "mkdev"
DATE = $(shell date '+%Y-%m-%d.%H:%M:%S')
IMAGE_VERSION ?= "v0.1.9" # TODO(mitchdraft) - replace with actual workflow

.PHONY: all
all: binaries containers ## (default) Builds binaries and containers

.PHONY: help
help:
	 @echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sort | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"

.PHONY: binaries
binaries: target/debugger-container/debugger-container target/squashctl # Builds squashctl binaries in and places them in target/ folder

RELEASE_BINARIES := target/squashctl-linux target/squashctl-osx

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: containers
containers: target/debugger-container-dlv-container target/debugger-container-gdb-container ## Builds debug containers

.PHONY: push-containers
push-containers: target/debugger-container-dlv-pushed target/debugger-container-gdb-pushed ## Pushes debug containers to $(DOCKER_REPO)

.PHONY: release
release: push-containers release-binaries ## Pushes containers to $(DOCKER_REPO) and releases binaries to GitHub

.PHONY: upload-release
upload-release: ## Uploads artifacts to GitHub releases
	./hack/github-release.sh owner=solo-io repo=squash tag=$(VERSION)
	@$(foreach BINARY,$(RELEASE_BINARIES),./hack/upload-github-release-asset.sh owner=solo-io repo=squash tag=$(VERSION) filename=$(BINARY);)

SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

# Pass in build-time variables
LDFLAGS := "-X github.com/solo-io/squash/pkg/version.Version=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.Timestamp=$(DATE) \
-X github.com/solo-io/squash/pkg/version.ImageVersion=$(IMAGE_VERSION) \
-X github.com/solo-io/squash/pkg/version.ImageRepo=$(DOCKER_REPO)"

target:
	[ -d $@ ] || mkdir -p $@

target/squashctl: target $(SRCS)
	go build -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl

target/squashctl-osx: target $(SRCS)
	GOOS=darwin go build -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl

target/squashctl-linux: target $(SRCS)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl

target/debugger-container/:
	[ -d $@ ] || mkdir -p $@

target/debugger-container/debugger-container: | target/debugger-container/
target/debugger-container/debugger-container: $(SRCS)
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o ./target/debugger-container/debugger-container ./cmd/debugger-container/


target/debugger-container/Dockerfile.dlv:    | target/debugger-container/
target/debugger-container/Dockerfile.dlv: cmd/debugger-container/Dockerfile.dlv
	cp cmd/debugger-container/Dockerfile.dlv target/debugger-container/Dockerfile.dlv
target/debugger-container-dlv-container: ./target/debugger-container/debugger-container target/debugger-container/Dockerfile.dlv
	docker build -f target/debugger-container/Dockerfile.dlv -t $(DOCKER_REPO)/debugger-container-dlv:$(VERSION) ./target/debugger-container/
	touch $@

target/debugger-container-dlv-pushed: target/debugger-container-dlv-container
	docker push $(DOCKER_REPO)/debugger-container-dlv:$(VERSION)
	touch $@



target/debugger-container/Dockerfile.gdb:    | target/debugger-container/
target/debugger-container/Dockerfile.gdb: cmd/debugger-container/Dockerfile.gdb
	cp cmd/debugger-container/Dockerfile.gdb target/debugger-container/Dockerfile.gdb
target/debugger-container-gdb-container: ./target/debugger-container/debugger-container target/debugger-container/Dockerfile.gdb
	docker build -f target/debugger-container/Dockerfile.gdb -t $(DOCKER_REPO)/debugger-container-gdb:$(VERSION) ./target/debugger-container/
	touch $@
target/debugger-container-gdb-pushed: target/debugger-container-gdb-container
	docker push $(DOCKER_REPO)/debugger-container-gdb:$(VERSION)
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
	jq '.version="$(VERSION)" | .binaries.linux="$(shell sha256sum target/squashctl-linux|cut -f1 -d" ")" | .binaries.darwin="$(shell sha256sum target/squashctl-osx|cut -f1 -d" ")"' src/squash.json > src/squash.json.tmp && \
	mv src/squash.json.tmp src/squash.json

.PHONY: clean
clean: ## Deletes target folder
	rm -rf target

dist: target/debugger-container-gdb-pushed target/debugger-container-dlv-pushed ## Pushes all containers to $(DOCKER_REPO)


## Temp
DEVVERSION="dev"
.PHONY: devpush
devpush:
	docker build -t $(DOCKER_REPO)/squash-agent:$(DEVVERSION) -f cmd/agent/Dockerfile ./target/agent/
	docker push $(DOCKER_REPO)/squash-agent:$(DEVVERSION)


# Docs

.PHONY: generatedocs
generatedocs:
	go run cmd/generate-docs/main.go
	mkdocs build

.PHONY: previewsite
previewsite:
	cd site && python3 -m http.server 0

####
# squashclt
# builds squashctl for each os (linux and darwin), puts output in the target/ dir
####

# .PHONY: squashctl
# squashctl: target/squashctl-osx target/squashctl-linux target/debugger-container/debugger-container
# 	echo "Building all squashctl"

# .PHONY: squashctldev
# squashctldev: target/squashctl target/debugger-container/debugger-container
# 	echo "just building for the default OS"

# # (convenience only) this one will will build squashctl for the builder's os
# target/squashctl: target $(SRCS)
# 	go build -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl/main.go

# target/squashctl-osx: target $(SRCS)
# 	GOOS=darwin go build -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl/main.go

# target/squashctl-linux: target $(SRCS)
# 	GOOS=linux go build -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl/main.go
