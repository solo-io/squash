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
/// Get the container object from its attachment.
type ContainerLocator interface {
	/// Takes a platform specific attachment object, and returns an updated attachment object, and a Container object.
	/// The updated attachment object will eventually make its way to ContainerProcess.GetContainerInfo
	Locate(context context.Context, attachment interface{}) (interface{}, *Container, error)
}

```

**Platforms** conform to interface used by squash client

```go

/// Runs in the squash client:
/// Information for the squash client to be able to connect debugger to the process
type ContainerInfo struct {
	Pid  int
	Name string
}

/// Get the information of a process that runs in the container. the pid should be in our pid namespace,
/// not in the container's namespace.
type ContainerProcess interface {
	/// Take a platform specific attachment object and return the pid the host pid namespace of the process we want to debug.
	GetContainerInfo(context context.Context, attachment interface{}) (*ContainerInfo, error)
}

```
