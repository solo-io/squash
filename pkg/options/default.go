package options

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

var (
	// The port where the debugger listens for remote connections
	// ( This is a port on the container that runs the debugger )
	DebuggerPort = 1235
	// OutPort is proxied by the debug-container process so that it can detect disconnections and terminate the debug session.
	// TODO(mitchdraft) - import this value from a common place (across squash and its IDE extensions)
	OutPort = 1236

	// The name used inside of a pod spec to refer to the container that runs the debugger
	PlankContainerName = "plank"

	// The root name (of the container image repo name) that will be shared among debugger-specific containers
	// Examples of particular container names: <RootName>-dlv, <RootName>-gdb, etc.
	ParticularContainerRootName = PlankContainerName

	SquashLabelSelectorKey   = "squash"
	SquashLabelSelectorValue = PlankContainerName
	PlankLabelSelectorString = fmt.Sprintf("%v=%v", SquashLabelSelectorKey, SquashLabelSelectorValue)

	// TODO(mitchdraft) - enable these debuggers
	// AvailableDebuggers = []string{"dlv", "gdb", "java", "java-port", "nodejs", "nodejs8", "python"}
	AvailableDebuggers = []string{"dlv", "java", "java-port", "gdb"}

	SquashPodName   = "squash"
	SquashNamespace = "squash-debugger"

	// squash permissions
	SquashServiceAccountName     = "squash"
	SquashClusterRoleName        = "squash-cr-pods"
	SquashClusterRoleBindingName = "squash-crb-pods"
	// optional secret for image pulls
	SquashServiceAccountImagePullSecretName = "squash-sa-image-pull-secret"

	// plank permissions
	PlankServiceAccountName     = "squash-plank"
	PlankClusterRoleName        = "squash-plank-cr"
	PlankClusterRoleBindingName = "squash-plank-crb"

	PlankEnvDebugAttachmentNamespace = "SQUASH_DEBUG_ATTACHMENT_NAMESPACE"
	PlankEnvDebugAttachmentName      = "SQUASH_DEBUG_ATTACHMENT_NAME"
	PlankEnvDebugSquashNamespace     = "SQUASH_DEBUG_SQUASH_NAMESPACE"

	KubeEnvPodName = "HOSTNAME"

	// This value is set in the Dockerfile
	PlankDockerEnvDebuggerType = "DEBUGGER"
)

// GeneratePlankLabels returns labels that  associate a plank with a given debug attachment
func GeneratePlankLabels(da *core.ResourceRef) map[string]string {
	labels := make(map[string]string)
	labels[SquashLabelSelectorKey] = SquashLabelSelectorValue
	labels["debug_attachment_namespace"] = da.Namespace
	labels["debug_attachment_name"] = da.Name
	labels[SquashLabelSelectorKey] = SquashLabelSelectorValue
	return labels
}
