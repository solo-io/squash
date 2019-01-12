package options

var (
	// In the future, this namespace will store a CRD that points to each of the active debugging attachments
	SquashCentralNamespace = "squash-debugger"

	// TODO(mitchdraft) - replace all occurances of SquashClientNamespace with the specific namespace of the pod that is being debugged
	SquashClientNamespace = SquashCentralNamespace
)
