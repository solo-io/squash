package options

import "fmt"

var (
	// The port where the debugger listens for remote connections
	// ( This is a port on the container that runs the debugger )
	DebuggerPort = 1235
	// OutPort is proxied by the debug-container process so that it can detect disconnections and terminate the debug session.
	// TODO(mitchdraft) - import this value from a common place (across squash and its IDE extensions)
	OutPort = 1236

	// The name used inside of a pod spec to refer to the container that runs the debugger
	ContainerName = "plank"

	// The root name (of the container image repo name) that will be shared among debugger-specific containers
	// Examples of particular container names: <RootName>-dlv, <RootName>-gdb, etc.
	ParticularContainerRootName = ContainerName

	SquashLabelSelectorKey    = "squash"
	SquashLabelSelectorValue  = ContainerName
	SquashLabelSelectorString = fmt.Sprintf("%v=%v", SquashLabelSelectorKey, SquashLabelSelectorValue)

	AvailableDebuggers = []string{"dlv", "gdb", "java", "nodejs", "nodejs8", "python"}

	SquashNamespace = "squash-debugger"

	PlankServiceAccountName     = "squash-plank"
	PlankClusterRoleName        = "squash-plank-cr"
	PlankClusterRoleBindingName = "squash-plank-crb"
)
