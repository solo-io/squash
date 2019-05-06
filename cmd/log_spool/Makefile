IMAGE_TAG ?= dev
CONTAINER_REPO_ORG ?= docker.io/soloio

## These are the images created by this makefile
APP_SPEC := $(CONTAINER_REPO_ORG)/log_spooler:$(IMAGE_TAG)

ROOTDIR := $(shell pwd)
OUTPUT_DIR := $(ROOTDIR)/_output

.PHONY: all
all: push-app

.PHONY: compile
compile: ${OUTPUT_DIR}
	GOOS=linux GOARCH=amd64 go build -gcflags "-N -l" -o $(OUTPUT_DIR)/log_spooler main.go

.PHONY: build-app
build-app: compile
	docker build -t $(APP_SPEC) -f Dockerfile .

.PHONY: push-app
push-app: build-app
	docker push $(APP_SPEC)
