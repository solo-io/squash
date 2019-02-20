#! /bin/bash -x

# this script is a helper for avoiding the trip to the container registry when iterating on processes that run in minikube
Set the below args as you need

DOCKER_REPO="soloio"
VERSION="mkdev2"
NS="demo"

eval $(minikube docker-env)

GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o ./target/debugger-container/debugger-container ./cmd/debugger-container/
docker build -f target/debugger-container/Dockerfile.dlv -t $DOCKER_REPO/debugger-container-dlv:$VERSION ./target/debugger-container/
docker build -f target/debugger-container/Dockerfile.gdb -t $DOCKER_REPO/debugger-container-gdb:$VERSION ./target/debugger-container/
