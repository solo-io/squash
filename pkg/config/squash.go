package config

type Squash struct {
	ChooseDebugger        bool
	NoClean               bool
	ChoosePod             bool
	TimeoutSeconds        int
	DebugContainerVersion string
	DebugContainerRepo    string
	LocalPort             int

	Debugger           string
	Namespace          string
	Pod                string
	Container          string
	Machine            bool
	DebugServerAddress string

	CRISock string
}
