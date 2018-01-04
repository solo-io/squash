# developer guide

Get Dev environment ready

[Prerequisites](#prerequisites)<BR>
[Platforms](#platforms)<BR>
[Code](#code)<BR>
[Build](#build)<BR>
[Package](#package)<BR>
[Deploy](#deploy)<BR>



## Prerequisites

* Working go environment. 
* [dep](https://github.com/golang/dep) tool for managing dependencies. <BR>
  Runs ```$ go get -u github.com/golang/dep/cmd/dep``` to install.
  
## Platforms  

* [Kubernetes](kubernetes.md)


## Code

Squash uses the new golang's dep tool for dependency management. to get the code and acquire dependencies, do:

```
$ mkdir -p $GOPATH/src/github.com/solo-io
$ cd $GOPATH/src/github.com/solo-io
$ git clone https://github.com/solo-io/squash
$ cd squash
$ dep ensure
```

## Build

To build the code, just do
```
$ make binaries
```

This will build the squash binaries.
## Package
To deploy squash to kubernetes we need to build containers images for use with kubernetes. 

To build the containers with a specific repository prefix, do:
```
$ make DOCKER_REPO=yourrepor VERSION=0.1
```
To push them using docker push, do:
```
$ make DOCKER_REPO=yourrepor VERSION=0.1  dist
```

## Deploy

You can use the `target/kubernetes/squash-client.yml` and `target/kubernetes/squash-server.yml` to deploy your containers (they will match the docker repo and version specificed in make).

Deploy (remember to deploy the server first):
```
kubectl create -f target/kubernetes/squash-server.yml
kubectl create -f target/kubernetes/squash-client.yml
```
# Directory structure
```
├── Gopkg.lock          <- Go dep's tool dependency lock file
├── Gopkg.toml          <- Go dep's tool config file
├── Makefile            <- Makefile to help build the project and containers
├── README.md
├── api.yaml            <- Squash API specification (Swagger 2.0)
├── cmd                 <- Command line tools
│  ├── squash-cli       <- The `squash` command line client
│  ├── squash-client    <- The squash client. Runs in every cluster node, and launces debug servers.
│  ├── squash-server    <- The squash server. Coordinates debug configs and sessions.  Auto generated. don't edit.
├── contrib             <- Additional scripts \ manifests
│  ├── example          <- Example to 2 simple microservices for demo\test purposes
│  └── kubernetes       <- Kubernetes deployment manifests for the squash server and client.
├── docs                <- Go here to learn more
└── pkg                 <- Most of the code goes here.
   ├── client           <- Auto generate REST client code. don't edit
   ├── debuggers        <- The debugger support in squash. see the "interface.go" file for more info.
   ├── models           <- Auto generate REST model code. don't edit
   ├── platforms        <- The platforms (clusters) support in squash. see the "interface.go" file for more info.
   ├── restapi          <- Auto generate REST code. don't edit
   ├── server           <- The implementation of the squash server is here.
   └── utils            <- Various utils (miscellanea).
```

# Contribution
We are looking for any contribution from the community in particular platforms, debuggers and IDEs.  
