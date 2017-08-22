package platforms

import "context"

/// Minimal represntation of a container, containing only the data squash cares about -
/// The container's name, image and the node it runs on.
type Container struct {
	Name, Image, Node string
}

/// Runs in the squash server:

/// Watch a service and get notifications when new containers of that service are created.
type ServiceWatcher interface {
	WatchService(context context.Context, servicename string) (<-chan Container, error)
}

/// Get the container object from its name.
// Note: in environment like kubernetes, the containername will be pod-name:container-name
type ContainerLocator interface {
	Locate(context context.Context, containername string) (*Container, error)
}

/// Runs in the squash client:

/// Get the pid of a process that runs in the container. the pid should be in our pid namespace,
/// not in the container's namespace.
type Container2Pid interface {
	GetPid(context context.Context, containername string) (int, error)
}
