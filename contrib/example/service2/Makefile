DOCKER_REPO ?= soloio
VERSION ?= dev
NS ?= demo

PODNAME=$(shell kubectl get pods -n demo --selector=app=example-service2 -o jsonpath='{.items[*].metadata.name}')

.PHONY=build
build:
	GOOS=linux CGO_ENABLED=0 go build -o service2 -gcflags "-N -l" main.go
	docker build -t $(DOCKER_REPO)/example-service2:$(VERSION) .

.PHONY=pushlocal
pushlocal: build
	kubectl apply -f service2dev.yml -n $(NS)
	kubectl delete pod -n $(NS) $(PODNAME)
