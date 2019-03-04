DOCKER_REPO ?= soloio
VERSION ?= v0.1.0

.PHONY: build
build:
	docker build -t $(DOCKER_REPO)/squash-demo-node-hello:$(VERSION) .

.PHONY: push
push:
	docker push $(DOCKER_REPO)/squash-demo-node-hello:$(VERSION)
