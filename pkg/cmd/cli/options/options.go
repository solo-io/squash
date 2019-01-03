package options

type Options struct {
	Url            string
	Json           bool
	DebugContainer DebugContainer
}

type DebugContainer struct {
	Name         string
	Namespace    string
	Image        string
	Pod          string
	Container    string
	ProcessName  string
	DebuggerType string
}
