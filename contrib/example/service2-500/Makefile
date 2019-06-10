DOCKER_REPO ?= soloio
VERSION ?= dev

.PHONY: build
build:
	GOOS=linux CGO_ENABLED=0 go build -o service2ise -gcflags "-N -l" main.go
	docker build -t $(DOCKER_REPO)/example-service2ise:$(VERSION) .

.PHONY: push
push: build
	docker push $(DOCKER_REPO)/example-service2ise:$(VERSION)

