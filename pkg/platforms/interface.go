package platforms

import (
	"context"

	v1 "github.com/solo-io/squash/pkg/api/v1"
)

/// Minimal represntation of a container, containing only the data squash cares about -
/// The container's name, image and the node it runs on.
type Container struct {
	Name, Image, Node string
}

/// Runs in the squash server:

/// Get the container object from its attachment.
type ContainerLocator interface {
	/// Takes a platform specific attachment object, and returns an updated attachment object, and a Container object.
	/// The updated attachment object will eventually make its way to ContainerProcess.GetContainerInfo
	Locate(context context.Context, attachment interface{}) (interface{}, *Container, error)
}

/// Runs in the squash client:
/// Information for the squash client to be able to connect debugger to the process
type ContainerInfo struct {
	Pids []int
	Name string
	Env  map[string]string
}

/// Get the information of a process that runs in the container. the pid should be in our pid namespace,
/// not in the container's namespace.
type ContainerProcess interface {
	/// Take a platform specific attachment object and return the pid the host pid namespace of the process we want to debug.
	GetContainerInfo(context context.Context, attachment *v1.DebugAttachment) (*ContainerInfo, error)
}
