#! /bin/bash -x -e

# Development helper script
# this script is a helper for avoiding the trip to the container registry when iterating on processes that run in minikube
# Set the below args as you need

DOCKER_REPO="soloio"
# VERSION="mkdev2"
VERSION="dev"

eval $(minikube docker-env)

echo "running command $1"
case $1 in
    "dc")
GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' -o ./target/plank/plank ./cmd/plank/
docker build -f target/plank/Dockerfile.dlv -t $DOCKER_REPO/plank-dlv:$VERSION ./target/plank/
docker build -f target/plank/Dockerfile.gdb -t $DOCKER_REPO/plank-gdb:$VERSION ./target/plank/
;;
    "agent")
# using Makefile to leverage its LDFLAGS spec
IMAGE_VERSION=$VERSION make tmpagent
docker build -t $DOCKER_REPO/squash-agent:$VERSION -f cmd/agent/Dockerfile ./target/agent/
;;
    *)
echo "unknown cmd" $1
;;
esac
