
REPO_PREFIX ?= docker.io
REPO_ORG ?= soloio
APP_NAME ?= squash-demo-cpp
VERSION ?= v0.0.1

IMAGE_SPEC := $(REPO_PREFIX)/$(REPO_ORG)/$(APP_NAME):$(VERSION)
.PHONY: build-push
build-push:
	docker build -t $(IMAGE_SPEC) . && docker push $(IMAGE_SPEC)

OUTPUT_DIR=_out
$(OUTPUT_DIR):
	mkdir $(OUTPUT_DIR)

# -g gives debug symbols
.PHONY: local
local: $(OUTPUT_DIR)
	g++ -std=c++11 -g -o $(OUTPUT_DIR)/$(APP_NAME) main.cpp
