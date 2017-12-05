package platforms

import "context"

/// Minimal represntation of a container, containing only the data squash cares about -
/// The container's name, image and the node it runs on.
type Container struct {
	Name, Image, Node string
}

/// Runs in the squash server:

/// Get the container object from its name.
// Note: in environment like kubernetes, the containername will be namespace:pod-name:container-name
type ContainerLocator interface {
	Locate(context context.Context, attachment interface{}) (interface{}, *Container, error)
}

/// Runs in the squash client:

/// Get the pid of a process that runs in the container. the pid should be in our pid namespace,
/// not in the container's namespace.
type Container2Pid interface {
	GetPid(context context.Context, attachment interface{}) (int, error)
}

type DataStore interface {
	Store()
	Load()
}
