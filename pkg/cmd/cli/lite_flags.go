package cli

import (
	"fmt"

	"github.com/solo-io/squash/pkg/kscmd"
	sqOpts "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/version"
	"github.com/spf13/pflag"
)

func applyLiteFlags(cfg *kscmd.SquashConfig, f *pflag.FlagSet) {
	f.BoolVar(&cfg.NoClean, "no-clean", false, "don't clean temporar pod when existing")
	f.BoolVar(&cfg.ChooseDebugger, "no-guess-debugger", false, "don't auto detect debugger to use")
	f.BoolVar(&cfg.ChoosePod, "no-guess-pod", false, "don't auto detect pod to use")
	f.BoolVar(&cfg.NoDetectSkaffold, "no-detect-pod", false, "don't auto settigns based on skaffold configuration present in current folder")
	f.BoolVar(&cfg.DebugServer, "debug-server", false, "start a debug server instead of an interactive session")
	f.IntVar(&cfg.TimeoutSeconds, "timeout", 300, "timeout in seconds to wait for debug pod to be ready")
	f.StringVar(&cfg.DebugContainerVersion, "container-version", version.ImageVersion, "debug container version to use")
	f.StringVar(&cfg.DebugContainerRepo, "container-repo", version.ImageRepo, "debug container repo to use")

	f.BoolVar(&cfg.LiteMode, "lite", true, "run in lite mode (default)")
	f.IntVar(&cfg.LocalPort, "localport", 0, fmt.Sprintf("port to use to connect to debugger (defaults to %v)", sqOpts.DebuggerPort))

	f.BoolVar(&cfg.Machine, "machine", false, "machine mode input and output")
	f.StringVar(&cfg.Debugger, "debugger", "dlv", "Debugger to use")
	f.StringVar(&cfg.Namespace, "namespace", "", "Namespace to debug")
	f.StringVar(&cfg.Pod, "pod", "", "Pod to debug")
	f.StringVar(&cfg.Container, "container", "", "Container to debug")
	f.StringVar(&cfg.CRISock, "crisock", "/var/run/dockershim.sock", "The path to the CRI socket")
}
