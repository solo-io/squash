DOCKER_REPO ?= soloio
VERSION ?= v0.2.2

service1-docker:
	cd service1 && docker run --rm -v $(PWD)/service1:$(PWD)/service1 -w $(PWD)/service1 -e GOOS=linux -e CGO_ENABLED=0 golang:1.8 go build -gcflags "-N -l"   && docker build --label io.solo.remotePath=$(PWD) -t $(DOCKER_REPO)/example-service1:$(VERSION) .
	cd service2 && docker run --rm -v $(PWD)/service2:$(PWD)/service2 -w $(PWD)/service2 -e GOOS=linux -e CGO_ENABLED=0 golang:1.8 go build -gcflags "-N -l"   && docker build --label io.solo.remotePath=$(PWD) -t $(DOCKER_REPO)/example-service2:$(VERSION) .
	cd service2-java && docker run --rm -v $(PWD)/service2-java:$(PWD)/service2-java -w $(PWD)/service2-java maven:3.5.2 mvn clean install && docker build --label io.solo.remotePath=$(PWD) -t $(DOCKER_REPO)/example-service2-java:$(VERSION) .

dist:
	docker push $(DOCKER_REPO)/example-service1:$(VERSION)
	docker push $(DOCKER_REPO)/example-service2:$(VERSION)
	docker push $(DOCKER_REPO)/example-service2-java:$(VERSION)