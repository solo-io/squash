# How to use Squash with istio

## Why?
Using Squash with istio enables you to debug the service mesh in a different way. Instead of attaching to a specific container you wish to debug, you (or the IDE for that matter) create a debug request that species which container image you want to debug. Then when a request is sent to the mesh with the "x-squash-debug" header, the mesh and Squash together will match the debug request to the specific container.

# Create environment

Note: To use Squash with istio we need Squash aware envoy and pilot. To make it easy, we have created an istio version with the Squash components baked inside. See istio.yaml in this folder. We will install it shown in the following steps.

## Kubernetes

Let's start by creating a kubernetes RBAC cluster. If you use minikube, these command can be helpful:

```
# create a cluster with a decent amount of resources, with RBAC enabled.
minikube  start --extra-config=apiserver.Authorization.Mode=RBAC --cpus 3 --memory 8192
# give kube dns the permissions it needs to work.
kubectl create clusterrolebinding permissive-binding   --clusterrole=cluster-admin   --user=admin   --user=kubelet   --group=system:serviceaccounts
```
Note: on Linux, if you don't have VirtualBox installed, you can try the `--vm-driver=kvm` option.

## Squash enabled Istio

Once the cluster is ready, install Squash aware Istio:

```
kubectl apply -f ./istio.yaml
```

Verify that everything is running, this may take a few minutes:
```
$ kubectl get pods --all-namespaces
NAMESPACE      NAME                             READY     STATUS    RESTARTS   AGE
istio-system   istio-ca-77ff87848b-ntwmj        1/1       Running   0          3m
istio-system   istio-ingress-5854489c59-wfqdl   1/1       Running   0          3m
istio-system   istio-mixer-5f7f4f696f-lwgwp     3/3       Running   0          3m
istio-system   istio-pilot-8459fdf5d8-hbgfp     2/2       Running   0          3m
kube-system    kube-addon-manager-minikube      1/1       Running   0          5m
kube-system    kube-dns-6fc954457d-phjtv        3/3       Running   0          4m
```

## Demo services

Deploy our demo services:
Note that if you use istioctl kube-inject, you need to modify them to use our version of the proxy container.
```
kubectl apply -f service2-istio.yaml
kubectl apply -f service1-istio.yaml
kubectl apply -f ingress.yaml
```

Get access to the ingress. for minikube users:
```
GATEWAY_HOST=$(kubectl get po -n istio-system -l istio=ingress -o 'jsonpath={.items[0].status.hostIP}')
GATEWAY_HTTP_INGRESS=$(kubectl get svc istio-ingress -n istio-system -o 'jsonpath={.spec.ports[?(@.name=="http")].nodePort}')
export HTTP_GATEWAY_URL=$GATEWAY_HOST:$GATEWAY_HTTP_INGRESS
```

## Install Squash

```
kubectl apply -f ./squash-server.yml
kubectl apply -f ./squash-client.yml
```


Verify that everything is running. This may take a few minutes.
```
$ kubectl get pods
NAME                                READY     STATUS    RESTARTS   AGE
example-service1-5bfd657c99-ltj45   2/2       Running   0          1m
example-service2-c94bf78bc-tlm4s    2/2       Running   0          1m
squash-client-49c8k                 1/1       Running   0          2m
squash-server-6cbf6cb8c7-26b4x      1/1       Running   0          2m
```

# Use the mesh to debug!

Once Squash is installed it's time to use the VSCode to debug the mesh.

To test that everything works as expect, try running `curl http://$HTTP_GATEWAY_URL/`. It should return an HTML page (you can also try open that URL in your browser).

## Accessing the cluster

To give the VSCode access to Squash server, please have `kubectl proxy` running in the background:
```
kubectl proxy&
```

Configure these settings in VSCode:
```
   "vs-squash.squash-path": "<PATH-TO-SQUASH-BINARY>/squash",
   "vs-squash.squash-server-url": "http://localhost:8001/api/v1/namespaces/default/services/squash-server:http-squash-api/proxy/api/v2" 
```
**Note:** that the Squash binary should be version 0.2 or higher.

## Prepare VSCode
For this tutorial, let's debug example-service1. Open VSCode in the `service1` folder. This is important for source code mapping (and consequently breakpoints) to work.
```
code ../example/service1
```
In the new VSCode window, place a breakpoint inside the handler method, around line 81.

In the new VSCode window, run the command: "Squash: Debug Container in Service Mesh"
Choose the image to debug by first selecting `example-service1` service and then the  `soloio/example-service1:v0.2.1`  image. The debugger you want to use is `dlv`. (Of course, make sure you have `dlv` installed and configured with VSCode Go extension).

VSCode will now be in a waiting mode. It will wait for a notification of a debug attachment.

## Trigger a debug enabled request

To trigger that debug attachment: Note the Squash debug header.
```
curl http://$HTTP_GATEWAY_URL/calc -H"x-squash-debug: true"
```

If all goes well, you will see the request hang and VSCode starts a debug session with the container that the request hit.
SUCCESS!

Note that you didn't have to specify the container yourself - the information about which container VSCode should attach to is provided by the Squash enabled service mesh.

## Notes:
You can use the command line to see whats going on. Specifically, try these commands:
```
squash list a --url=http://localhost:8001/api/v1/namespaces/default/services/squash-server:http-squash-api/proxy/api/v2
squash list r --url=http://localhost:8001/api/v1/namespaces/default/services/squash-server:http-squash-api/proxy/api/v2
```

# More Resources:
- https://github.com/solo-io/istio-squash
- https://github.com/solo-io/envoy-squash
