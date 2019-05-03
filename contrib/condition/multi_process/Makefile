IMAGE_TAG ?= dev
CONTAINER_REPO_ORG ?= docker.io/soloio

## These are the images created by this makefile
MULTI_SPEC := $(CONTAINER_REPO_ORG)/multi_process:$(IMAGE_TAG)
SINGLE_SPEC := $(CONTAINER_REPO_ORG)/multi_process_base:$(IMAGE_TAG)

ROOTDIR := $(shell pwd)
OUTPUT_DIR := $(ROOTDIR)/_output

.PHONY: all
all: push-single push-multi

.PHONY: compile
compile:
	GOOS=linux go build -gcflags "-N -l" -o $(OUTPUT_DIR)/sample_app main.go

.PHONY: build-multi
build-multi: compile
	docker build -t $(MULTI_SPEC) -f multi.dockerfile .

.PHONY: push-multi
push-multi: build-multi
	docker push $(MULTI_SPEC)

.PHONY: build-single
build-single: compile
	docker build -t $(SINGLE_SPEC) -f single.dockerfile .

.PHONY: push-single
push-single: build-single
	docker push $(SINGLE_SPEC)
