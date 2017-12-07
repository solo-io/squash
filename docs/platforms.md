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
	/// The updated attachment object will eventually make its way to Container2Pid.GetPid
	Locate(context context.Context, attachment interface{}) (interface{}, *Container, error)
}

```

**Platforms** conform to interface used by squash client

```go
/// Get the pid of a process that runs in the container. the pid should be in our pid namespace,
/// not in the container's namespace.
type Container2Pid interface {
	/// Take a platform specific attachment object and return the pid the host pid namespace of the process we want to debug.
	GetPid(context context.Context, attachment interface{}) (int, error)
}

```
