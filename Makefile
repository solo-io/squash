#----------------------------------------------------------------------------------
# This portion is managed by github.com/solo-io/build
#----------------------------------------------------------------------------------
# NOTE! All make targets that use the computed values must depend on the "must"
# target to ensure the expected computed values were recieved
.PHONY: must
must: validate-computed-values

# Read computed values into variables that can be used by make
# Since both stdout and stderr are passed, our make targets validate the variables
RELEASE := $(shell go run buildcmd/main.go parse-env release)
VERSION := $(shell go run buildcmd/main.go parse-env version)
IMAGE_TAG := $(shell go run buildcmd/main.go parse-env image-tag)
CONTAINER_REPO_ORG := $(shell go run buildcmd/main.go parse-env container-prefix)

# use this, or the shorter alias "must", as a dependency for any target that uses
# values produced by the build tool
.PHONY: validate-computed-values
validate-computed-values:
	go run buildcmd/main.go validate-operating-parameters \
		$(RELEASE) \
		$(VERSION) \
		$(CONTAINER_REPO_ORG) \
		$(IMAGE_TAG)

.PHONY: preview-computed-values
preview-computed-values: must
	echo summary of computed values - \
		release: $(RELEASE), \
		version: $(VERSION), \
		container-prefix: $(CONTAINER_REPO_ORG), \
		image-tag: $(IMAGE_TAG)

#### END OF MANAGED PORTION

#----------------------------------------------------------------------------------
# Base
#----------------------------------------------------------------------------------

ROOTDIR := $(shell pwd)
OUTPUT_DIR := $(ROOTDIR)/_output
DATE = $(shell date '+%Y-%m-%d.%H:%M:%S')
SRCS=$(shell find ./pkg -name "*.go") $(shell find ./cmd -name "*.go")

# Pass in build-time variables
LDFLAGS := "-X github.com/solo-io/squash/pkg/version.Version=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.TimeStamp=$(DATE) \
-X github.com/solo-io/squash/pkg/version.ImageVersion=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.SquashImageTag=$(VERSION) \
-X github.com/solo-io/squash/pkg/version.ImageRepo=$(CONTAINER_REPO_ORG)"

.PHONY: all
all: must release-binaries containers ## (default) Builds binaries and containers

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
generatecode: must
	mkdir -p $(OUTPUT_DIR)
	go run cmd/generate-code/main.go
	rm docs/cli/squashctl*
	go run cmd/generate-docs/main.go
	gofmt -w ci cmd pkg test
	goimports -w ci cmd pkg test

# for use by ci
# if any docs have changed, this will create a PR on the solo-io/solo-docs repo
# assumes TAGGED_VERSION and GITHUB_TOKEN are in env
.PHONY: push-docs
push-docs: must
	go run ci/push_docs.go

#----------------------------------------------------------------------------------
# Squashctl
#----------------------------------------------------------------------------------
.PHONY: squashctl
squashctl: must $(OUTPUT_DIR)/squashctl $(OUTPUT_DIR)/squashctl-darwin $(OUTPUT_DIR)/squashctl-linux $(OUTPUT_DIR)/squashctl-windows.exe

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
squash: must $(OUTPUT_DIR)/squash-container

$(OUTPUT_DIR)/squash: $(SRCS)
	GOOS=linux go build -ldflags=$(LDFLAGS) -o $(OUTPUT_DIR)/squash/squash cmd/squash/main.go
$(OUTPUT_DIR)/squash-container: $(OUTPUT_DIR)/squash
	docker build -f cmd/squash/Dockerfile -t $(CONTAINER_REPO_ORG)/squash:$(IMAGE_TAG) $(OUTPUT_DIR)/squash/
	touch $@

#----------------------------------------------------------------------------------
# Plank
#----------------------------------------------------------------------------------
.PHONY: plank
plank: must $(OUTPUT_DIR)/plank-dlv-container $(OUTPUT_DIR)/plank-gdb-container

$(OUTPUT_DIR)/plank/:
	[ -d $@ ] || mkdir -p $@

$(OUTPUT_DIR)/plank/plank: | $(OUTPUT_DIR)/plank/
$(OUTPUT_DIR)/plank/plank: $(SRCS)
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -ldflags=$(LDFLAGS) -o $(OUTPUT_DIR)/plank/plank ./cmd/plank/

$(OUTPUT_DIR)/plank/Dockerfile.dlv:    | $(OUTPUT_DIR)/plank/
$(OUTPUT_DIR)/plank/Dockerfile.dlv: cmd/plank/Dockerfile.dlv
	cp cmd/plank/Dockerfile.dlv $(OUTPUT_DIR)/plank/Dockerfile.dlv
$(OUTPUT_DIR)/plank-dlv-container: $(OUTPUT_DIR)/plank/plank $(OUTPUT_DIR)/plank/Dockerfile.dlv
	docker build -f $(OUTPUT_DIR)/plank/Dockerfile.dlv -t $(CONTAINER_REPO_ORG)/plank-dlv:$(IMAGE_TAG) $(OUTPUT_DIR)/plank/
	touch $@

$(OUTPUT_DIR)/plank/Dockerfile.gdb:    | $(OUTPUT_DIR)/plank/
$(OUTPUT_DIR)/plank/Dockerfile.gdb: cmd/plank/Dockerfile.gdb
	cp cmd/plank/Dockerfile.gdb $(OUTPUT_DIR)/plank/Dockerfile.gdb
$(OUTPUT_DIR)/plank-gdb-container: $(OUTPUT_DIR)/plank/plank $(OUTPUT_DIR)/plank/Dockerfile.gdb
	docker build -f $(OUTPUT_DIR)/plank/Dockerfile.gdb -t $(CONTAINER_REPO_ORG)/plank-gdb:$(IMAGE_TAG) $(OUTPUT_DIR)/plank/
	touch $@

#----------------------------------------------------------------------------------
# VS-Code extension
#----------------------------------------------------------------------------------
.PHONY: publish-extension
publish-extension: must package-extension ## (vscode) Publishes extension
ifeq ($(RELEASE),TRUE)
	./hack/publish-extension.sh
	touch $@
endif

.PHONY: package-extension
package-extension: must bump-extension-version ## (vscode) Packages extension
ifeq ($(RELEASE),TRUE)
	cd editor/vscode && npm install --unsafe-perm
	cd editor/vscode && vsce package
	touch $@
endif

.PHONY: bump-extension-version
bump-extension-version: must  ## (vscode) Bumps extension version
ifeq ($(RELEASE),TRUE)
	go run ci/bump_extension_version.go $(VERSION)
	touch $@
endif

#----------------------------------------------------------------------------------
# Deployment Manifests / Helm
#----------------------------------------------------------------------------------

HELM_SYNC_DIR := $(OUTPUT_DIR)/helm
HELM_DIR := install/helm/$(SOLO_NAME)
INSTALL_NAMESPACE ?= $(SOLO_NAME)

.PHONY: manifest
manifest: must prepare-helm install/squash.yaml update-helm-chart

# creates Chart.yaml, values.yaml
.PHONY: prepare-helm
prepare-helm: must
	go run install/helm/squash/generate/cmd/generate.go $(IMAGE_TAG) $(CONTAINER_REPO_ORG)

.PHONY: update-helm-chart
update-helm-chart: must
	mkdir -p $(HELM_SYNC_DIR)/charts
	helm package --destination $(HELM_SYNC_DIR)/charts $(HELM_DIR)
	helm repo index $(HELM_SYNC_DIR)

HELMFLAGS := --namespace $(INSTALL_NAMESPACE) --set namespace.create=true

install/$(SOLO_NAME).yaml: prepare-helm
	helm template install/helm/squash $(HELMFLAGS) > $@

.PHONY: render-yaml
render-yaml: must install/squash.yaml

#----------------------------------------------------------------------------------
# Build All
#----------------------------------------------------------------------------------
.PHONY: build
build: must squashctl squash plank

#----------------------------------------------------------------------------------
# Docker
#----------------------------------------------------------------------------------
.PHONY: docker
docker: must $(OUTPUT_DIR)/plank-dlv-container $(OUTPUT_DIR)/plank-gdb-container $(OUTPUT_DIR)/squash-container

.PHONY: docker-push
docker-push: must docker
	docker push $(CONTAINER_REPO_ORG)/plank-dlv:$(IMAGE_TAG) && \
	docker push $(CONTAINER_REPO_ORG)/plank-gdb:$(IMAGE_TAG) && \
	docker push $(CONTAINER_REPO_ORG)/squash:$(IMAGE_TAG)

#----------------------------------------------------------------------------------
# Release
#----------------------------------------------------------------------------------
.PHONY: upload-github-release-assets
upload-github-release-assets: must squashctl
	go run ci/upload_github_release_assets.go

#----------------------------------------------------------------------------------
# Development utils
#----------------------------------------------------------------------------------
# Helpers for development: build and push (locally) only the things you changed
# first run `eval $(minikube docker-env)` then any of these commands
.PHONY: dev-squashctl-darwin
dev-squashctl-darwin: must $(OUTPUT_DIR) $(SRCS) $(OUTPUT_DIR)/squashctl-darwin

.PHONY: dev-squashctl-win
dev-squashct-win: must $(OUTPUT_DIR)/squashctl-windows

.PHONY: dev-planks
dev-planks: must $(OUTPUT_DIR) $(SRCS) $(OUTPUT_DIR)/plank-dlv-container $(OUTPUT_DIR)/plank-gdb-container

.PHONY: dev-squash
dev-planks: must $(OUTPUT_DIR) $(SRCS) $(OUTPUT_DIR)/squash-container
