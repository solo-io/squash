#! /bin/bash -x

DOCKER_REPO="soloio"
VERSION="mkdev2"
NS="t2"
PODNAME=$(kubectl get pods -n $NS --selector=app=example-service2-java -o jsonpath='{.items[*].metadata.name}')

eval $(minikube docker-env)

echo $(PWD)

cd service2-java && docker run --rm -v $(PWD)/service2-java:$(PWD)/service2-java -w $(PWD)/service2-java maven:3.5.2 mvn clean install && docker build --label io.solo.remotePath=$(PWD) -t $DOCKER_REPO/example-service2-java:$VERSION .
docker build -t $DOCKER_REPO/example-service2-java:$VERSION .

kubectl apply -f service2dev-java.yml -n $NS
kubectl delete pod -n $NS $PODNAME
