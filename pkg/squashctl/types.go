package squashctl

import (
	"context"

	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/config"
	"k8s.io/client-go/kubernetes"
)

type Options struct {
	KubeClient *kubernetes.Clientset

	Url            string
	Json           bool
	DebugContainer DebugContainer
	// Debug Container is a superset of DebugRequest so we can use the same struct
	// TODO(mitchdraft) - refactor
	DebugRequest DebugContainer
	daClient     v1.DebugAttachmentClient
	ctx          context.Context
	Wait         Wait

	Squash      config.Squash
	DebugTarget config.DebugTarget

	DeployOptions DeployOptions

	// RbacMode bool

	// Verbose controls how much contextual information is printed
	// Verbose bool

	// Internal contains cli-specific metadata
	Internal Internal

	// Config may be blended into other options
	Config Config
}

func NewOptions() *Options {
	o := &Options{}
	o.Squash = config.NewSquashConfig()
	return o
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

type Wait struct {
	Timeout float64
}

type Error struct {
	Type string
	Info string
}

type DeployOptions struct {
	DemoOptions          DemoOptions
	SquashProcessOptions SquashProcessOptions
}

func defaultDeployOptions() DeployOptions {
	return DeployOptions{
		DemoOptions:          defaultDemoOptions(),
		SquashProcessOptions: defaultSquashProcessOptions(),
	}
}

type DemoOptions struct {
	Namespace1 string
	Namespace2 string
	DemoId     string
}

func defaultDemoOptions() DemoOptions {
	return DemoOptions{
		Namespace1: "squash",
		Namespace2: "squash",
		DemoId:     "go-go",
	}
}

type SquashProcessOptions struct {
	Namespace string
	// Preview, if set prints a yaml description of the Squash installation without creating installing Squash
	Preview bool
}

func defaultSquashProcessOptions() SquashProcessOptions {
	return SquashProcessOptions{
		Namespace: "squash-debugger",
	}
}

type Internal struct {
	// ConfigLoaded should be set once the config has been loaded
	ConfigLoaded bool
	// ConfigRead should be set once the config has been read
	ConfigRead bool
}

type Config struct {
	verbose    bool
	secureMode bool
	logCmds    bool

	// values from the kubesquash extension that have not been implemented
	// path to the preferred kubectl binary
	kubectlPath string // NOT IMPLEMENTED
	// arguments to be passed to kubectl via `KUBECONFIG=<kubeEnv> kubectl ...`
	kubeEnv string // NOT IMPLEMENTED
	// arguments to be passed to kubectl via ` kubectl --kubeconfig=<kubeConfig> ...`
	kubeConfig string // NOT IMPLEMENTED
	// semi-implemented as flags, should move to config file
	containerRepo // SEMI-IMPLEMENTED

}
