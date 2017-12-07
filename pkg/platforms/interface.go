package platforms

import "context"

/// Minimal represntation of a container, containing only the data squash cares about -
/// The container's name, image and the node it runs on.
type Container struct {
	Name, Image, Node string
}

/// Runs in the squash server:

/// Get the container object from its attachment.
type ContainerLocator interface {
	/// Takes a platform specific attachment object, and returns an updated attachment object, and a Container object.
	/// The updated attachment object will eventually make its way to Container2Pid.GetPid
	Locate(context context.Context, attachment interface{}) (interface{}, *Container, error)
}

/// Runs in the squash client:

/// Get the pid of a process that runs in the container. the pid should be in our pid namespace,
/// not in the container's namespace.
type Container2Pid interface {
	/// Take a platform specific attachment object and return the pid the host pid namespace of the process we want to debug.
	GetPid(context context.Context, attachment interface{}) (int, error)
}