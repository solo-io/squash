DOCKER_REPO ?= soloio
DATE = $(shell date '+%Y-%m-%d.%H:%M:%S')

# produce a release if TAGGED_VERSION is set
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	TAGGED_VERSION := vdev
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

.PHONY: all
all: binaries containers ## (default) Builds binaries and containers

.PHONY: help
help:
	 @echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sort | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"

.PHONY: binaries
binaries: target/plank/plank target/squashctl target/squash # Builds squashctl binaries in and places them in target/ folder

RELEASE_BINARIES := target/squashctl-linux target/squashctl-osx

.PHONY: release-binaries
release-binaries: $(RELEASE_BINARIES)

.PHONY: containers
containers: target/plank-dlv-container target/plank-gdb-container ## Builds debug containers

.PHONY: push-containers
push-containers: all target/plank-dlv-pushed target/plank-gdb-pushed target/squash-pushed ## Pushes debug containers to $(DOCKER_REPO)

.PHONY: release
release: push-containers release-binaries ## Pushes containers to $(DOCKER_REPO) and releases binaries to GitHub

.PHONY: upload-release
upload-release: ## Uploads artifacts to GitHub releases
	./hack/github-release.sh owner=solo-io repo=squash tag=$(VERSION)
	@$(foreach BINARY,$(RELEASE_BINARIES),./hack/upload-github-release-asset.sh owner=solo-io repo=squash tag=$(VERSION) filename=$(BINARY);)

SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

# Pass in build-time variables
LDFLAGS := "-X github.com/solo-io/squash/pkg/version.Version=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.TimeStamp=$(DATE) \
-X github.com/solo-io/squash/pkg/version.ImageVersion=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.SquashImageTag=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.ImageRepo=$(DOCKER_REPO)"

target:
	[ -d $@ ] || mkdir -p $@

### Squashctl


target/squashctl-osx: target $(SRCS) target/squashctl
	GOOS=darwin go build -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl

target/squashctl-linux: target $(SRCS)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl


### Squash

target/squash: target $(SRCS) generatecode
	GOOS=linux go build -ldflags=$(LDFLAGS) -o target/squash/squash cmd/squash/main.go

target/squash-container: ./target/squash
	docker build -f cmd/squash/Dockerfile -t $(DOCKER_REPO)/squash:$(VERSION) ./target/squash/
	touch $@
target/squash-pushed: target/squash-container
	docker push $(DOCKER_REPO)/squash:$(VERSION)
	touch $@


 ### Plank

target/plank/: generatecode
	[ -d $@ ] || mkdir -p $@

target/plank/plank: | target/plank/
target/plank/plank: $(SRCS)
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o ./target/plank/plank ./cmd/plank/


target/plank/Dockerfile.dlv:    | target/plank/
target/plank/Dockerfile.dlv: cmd/plank/Dockerfile.dlv
	cp cmd/plank/Dockerfile.dlv target/plank/Dockerfile.dlv
target/plank-dlv-container: ./target/plank/plank target/plank/Dockerfile.dlv
	docker build -f target/plank/Dockerfile.dlv -t $(DOCKER_REPO)/plank-dlv:$(VERSION) ./target/plank/
	touch $@

target/plank-dlv-pushed: target/plank-dlv-container
	docker push $(DOCKER_REPO)/plank-dlv:$(VERSION)
	touch $@



target/plank/Dockerfile.gdb:    | target/plank/
target/plank/Dockerfile.gdb: cmd/plank/Dockerfile.gdb
	cp cmd/plank/Dockerfile.gdb target/plank/Dockerfile.gdb
target/plank-gdb-container: ./target/plank/plank target/plank/Dockerfile.gdb
	docker build -f target/plank/Dockerfile.gdb -t $(DOCKER_REPO)/plank-gdb:$(VERSION) ./target/plank/
	touch $@
target/plank-gdb-pushed: target/plank-gdb-container
	docker push $(DOCKER_REPO)/plank-gdb:$(VERSION)
	touch $@

.PHONY: publish-extension
publish-extension: bump-extension-version ## (vscode) Publishes extension
	./hack/publish-extension.sh

.PHONY: package-extension
package-extension: bump-extension-version ## (vscode) Packages extension
	cd editor/vscode && vsce package

.PHONY: bump-extension-version
bump-extension-version:  ## (vscode) Bumps extension version
	cd editor/vscode && \
	jq '.version="$(VERSION)" | .version=.version[1:]' package.json > package.json.tmp && \
	mv package.json.tmp package.json && \
	jq '.version="$(VERSION)" | .binaries.linux="$(shell sha256sum target/squashctl-linux|cut -f1 -d" ")" | .binaries.darwin="$(shell sha256sum target/squashctl-osx|cut -f1 -d" ")"' src/squash.json > src/squash.json.tmp && \
	mv src/squash.json.tmp src/squash.json

.PHONY: clean
clean: ## Deletes target folder
	rm -rf target

dist: target/plank-gdb-pushed target/plank-dlv-pushed ## Pushes all containers to $(DOCKER_REPO)

# Generated code
.PHONY: generatecode
generatecode:
	go run cmd/generate-code/main.go

# Docs

.PHONY: generatedocs
generatedocs:
	go run cmd/generate-docs/main.go
	mkdocs build

.PHONY: previewsite
previewsite:
	cd site && python3 -m http.server 0


# Helpers for development: build and push (locally) only the things you changed
# first run `eval $(minikube docker-env)` then any of these commands
.PHONY: dev_squashctl
dev_squashctl: target $(SRCS) target/squashctl

.PHONY: dev_planks
dev_planks: target $(SRCS) target/plank-dlv-container target/plank-gdb-container

.PHONY: dev_squash
dev_planks: target $(SRCS) target/squash-container
