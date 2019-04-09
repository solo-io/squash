#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR ?= $(ROOTDIR)/_output
DOCKER_REPO ?= quay.io/solo-io
DATE = $(shell date '+%Y-%m-%d.%H:%M:%S')
SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

# produce a release if TAGGED_VERSION is set
RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	TAGGED_VERSION := vdev
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

# Pass in build-time variables
LDFLAGS := "-X github.com/solo-io/squash/pkg/version.Version=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.TimeStamp=$(DATE) \
-X github.com/solo-io/squash/pkg/version.ImageVersion=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.SquashImageTag=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.ImageRepo=$(DOCKER_REPO)"

.PHONY: all
all: release-binaries containers ## (default) Builds binaries and containers

.PHONY: help
help:
	 @echo -e "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sort | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s :)"

#----------------------------------------------------------------------------------
# Repo setup
#----------------------------------------------------------------------------------

# https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/
.PHONY: init
init:
	git config core.hooksPath .githooks

.PHONY: update-deps
update-deps:
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/gogo/protobuf/gogoproto
	go get -u github.com/gogo/protobuf/protoc-gen-gogo
	mkdir -p $$GOPATH/src/github.com/lyft
	cd $$GOPATH/src/github.com/lyft && if [ ! -e protoc-gen-validate ];then git clone https://github.com/envoyproxy/protoc-gen-validate; fi && cd protoc-gen-validate && git checkout v0.0.6
	go get -u github.com/paulvollmer/2gobytes

.PHONY: pin-repos
pin-repos:
	go run ci/pin_repos.go

.PHONY: check-format
check-format:
	NOT_FORMATTED=$$(gofmt -l ./pkg/ ./cmd/ ) && if [ -n "$$NOT_FORMATTED" ]; then echo These files are not formatted: $$NOT_FORMATTED; exit 1; fi

#----------------------------------------------------------------------------------
# Clean
#----------------------------------------------------------------------------------

# Important to clean before pushing new releases. Dockerfiles and binaries may not update properly
.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)
	rm -rf site


#----------------------------------------------------------------------------------
# Generated Code and Docs
#----------------------------------------------------------------------------------
# Generated code
.PHONY: generatecode
generatecode:
	mkdir -p $(OUTPUT_DIR)
	go run cmd/generate-code/main.go
	# go run cmd/generate-docs/main.go # Re-enable this when squashctl does not create resources on init
	gofmt -w ci cmd pkg test
	goimports -w ci cmd pkg test

# for use by ci
# if any docs have changed, this will create a PR on the solo-io/solo-docs repo
# assumes TAGGED_VERSION and GITHUB_TOKEN are in env
.PHONY: push-docs
push-docs:
	go run ci/push_docs.go


#----------------------------------------------------------------------------------
# Squashctl
#----------------------------------------------------------------------------------
.PHONY: squashctl
squashctl: $(OUTPUT_DIR)/squashctl $(OUTPUT_DIR)/squashctl-darwin $(OUTPUT_DIR)/squashctl-linux $(OUTPUT_DIR)/squashctl-windows.exe

$(OUTPUT_DIR)/squashctl: $(SRCS)
	go build -a -tags netgo -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl

$(OUTPUT_DIR)/squashctl-darwin: $(SRCS)
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl
$(OUTPUT_DIR)/squashctl-linux: $(SRCS)
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl
$(OUTPUT_DIR)/squashctl-windows.exe: $(SRCS)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o $@ ./cmd/squashctl


#----------------------------------------------------------------------------------
# Squash
#----------------------------------------------------------------------------------
.PHONY: squash
squash: $(OUTPUT_DIR)/squash-container

$(OUTPUT_DIR)/squash: $(SRCS)
	GOOS=linux go build -ldflags=$(LDFLAGS) -o $(OUTPUT_DIR)/squash/squash cmd/squash/main.go
$(OUTPUT_DIR)/squash-container: $(OUTPUT_DIR)/squash
	docker build -f cmd/squash/Dockerfile -t $(DOCKER_REPO)/squash:$(VERSION) $(OUTPUT_DIR)/squash/
	touch $@


#----------------------------------------------------------------------------------
# Plank
#----------------------------------------------------------------------------------
.PHONY: plank
plank: $(OUTPUT_DIR)/plank-dlv-container $(OUTPUT_DIR)/plank-gdb-container

$(OUTPUT_DIR)/plank/:
	[ -d $@ ] || mkdir -p $@

$(OUTPUT_DIR)/plank/plank: | $(OUTPUT_DIR)/plank/
$(OUTPUT_DIR)/plank/plank: $(SRCS)
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o $(OUTPUT_DIR)/plank/plank ./cmd/plank/

$(OUTPUT_DIR)/plank/Dockerfile.dlv:    | $(OUTPUT_DIR)/plank/
$(OUTPUT_DIR)/plank/Dockerfile.dlv: cmd/plank/Dockerfile.dlv
	cp cmd/plank/Dockerfile.dlv $(OUTPUT_DIR)/plank/Dockerfile.dlv
$(OUTPUT_DIR)/plank-dlv-container: $(OUTPUT_DIR)/plank/plank $(OUTPUT_DIR)/plank/Dockerfile.dlv
	docker build -f $(OUTPUT_DIR)/plank/Dockerfile.dlv -t $(DOCKER_REPO)/plank-dlv:$(VERSION) $(OUTPUT_DIR)/plank/
	touch $@

$(OUTPUT_DIR)/plank/Dockerfile.gdb:    | $(OUTPUT_DIR)/plank/
$(OUTPUT_DIR)/plank/Dockerfile.gdb: cmd/plank/Dockerfile.gdb
	cp cmd/plank/Dockerfile.gdb $(OUTPUT_DIR)/plank/Dockerfile.gdb
$(OUTPUT_DIR)/plank-gdb-container: $(OUTPUT_DIR)/plank/plank $(OUTPUT_DIR)/plank/Dockerfile.gdb
	docker build -f $(OUTPUT_DIR)/plank/Dockerfile.gdb -t $(DOCKER_REPO)/plank-gdb:$(VERSION) $(OUTPUT_DIR)/plank/
	touch $@


#----------------------------------------------------------------------------------
# VS-Code extension
#----------------------------------------------------------------------------------
.PHONY: publish-extension
publish-extension: package-extension ## (vscode) Publishes extension
ifeq ($(RELEASE),"true")
	./hack/publish-extension.sh
	touch $@
endif

.PHONY: package-extension
package-extension: bump-extension-version ## (vscode) Packages extension
ifeq ($(RELEASE),"true")
	cd editor/vscode && npm install --unsafe-perm
	cd editor/vscode && vsce package
	touch $@
endif

.PHONY: bump-extension-version
bump-extension-version:  ## (vscode) Bumps extension version
	go run ci/bump_extension_version.go $(VERSION)
	touch $@


#----------------------------------------------------------------------------------
# Build All
#----------------------------------------------------------------------------------
.PHONY: build
build: squashctl squash plank


#----------------------------------------------------------------------------------
# Docker
#----------------------------------------------------------------------------------
.PHONY: docker
docker: $(OUTPUT_DIR)/plank-dlv-container $(OUTPUT_DIR)/plank-gdb-container $(OUTPUT_DIR)/squash-container

DOCKER_IMAGES :=
ifeq ($(RELEASE),"true")
	DOCKER_IMAGES := docker
endif

.PHONY: docker-push
docker-push: $(DOCKER_IMAGES)
ifeq ($(RELEASE),"true")
	docker push $(DOCKER_REPO)/plank-dlv:$(VERSION) && \
	docker push $(DOCKER_REPO)/plank-gdb:$(VERSION) && \
	docker push $(DOCKER_REPO)/squash:$(VERSION)
endif


#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------
.PHONY: upload-github-release-assets
upload-github-release-assets: squashctl
	go run ci/upload_github_release_assets.go


#----------------------------------------------------------------------------------
# Development utils
#----------------------------------------------------------------------------------
# Helpers for development: build and push (locally) only the things you changed
# first run `eval $(minikube docker-env)` then any of these commands
.PHONY: dev-squashctl-darwin
dev-squashctl-darwin: $(OUTPUT_DIR) $(SRCS) $(OUTPUT_DIR)/squashctl-darwin

.PHONY: dev-squashctl-win
dev-squashct-win: $(OUTPUT_DIR)/squashctl-windows

.PHONY: dev-planks
dev-planks: $(OUTPUT_DIR) $(SRCS) $(OUTPUT_DIR)/plank-dlv-container $(OUTPUT_DIR)/plank-gdb-container

.PHONY: dev-squash
dev-planks: $(OUTPUT_DIR) $(SRCS) $(OUTPUT_DIR)/squash-container
