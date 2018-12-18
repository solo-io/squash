# Kubernetes

Squash requires kubernetes version 1.7 and up.

Squash also works with [minikube](https://kubernetes.io/docs/getting-started-guides/minikube/).

## Prerequisites:
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) that's pre-configured for your cluster.


## Install
Execute in order:
```
$ kubectl create -f https://raw.githubusercontent.com/solo-io/squash/master/contrib/kubernetes/squash-server.yml
$ kubectl create -f https://raw.githubusercontent.com/solo-io/squash/master/contrib/kubernetes/squash-client.yml
```

The next step is to [get started](../getting-started.md) to debug your first microservice with squash. 
