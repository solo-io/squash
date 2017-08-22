# Platforms

Currently Squash support the following platformrs:
  - [Kubernetes](docs/platforms/kubernetes.md)

Future planes:
  - [Mesos](http://mesos.apache.org)
  - [Docker Swam](https://github.com/docker/swarm)
  - [Cloud Foundry](https://www.cloudfoundry.org)

<BR>
=> We are looking for community help to add support for more Platforms.

##

**Platforms** conform to interfaces used by squash server

```go

type ServiceWatcher interface {

	WatchService(context context.Context, servicename string) (<-chan Container, error)
}

type ContainerLocator interface {

	Locate(context context.Context, containername string) (*Container, error)
}
```

**Platforms** conform to interface used by squash client

```go
type Container2Pid interface {

	GetPid(context context.Context, containername string) (int, error)
}
```
