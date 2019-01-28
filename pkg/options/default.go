package options

import "fmt"

var (
	// The port where the debugger listens for remote connections
	// ( This is a port on the container that runs the debugger )
	DebuggerPort = "1235"

	// The name used inside of a pod spec to refer to the container that runs the debugger
	ContainerName = "kubesquash-container"

	// The root name (of the container image repo name) that will be shared among debugger-specific containers
	// Examples of particular container names: <RootName>-dlv, <RootName>-gdb, etc.
	ParticularContainerRootName = ContainerName

	SquashLabelSelectorKey    = "squash"
	SquashLabelSelectorValue  = ContainerName
	SquashLabelSelectorString = fmt.Sprintf("%v=%v", SquashLabelSelectorKey, SquashLabelSelectorValue)
)
