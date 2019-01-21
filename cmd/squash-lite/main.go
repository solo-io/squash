package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/solo-io/squash/pkg/lite/kube"
)

const descriptionUsage = `Normally squash lite requires no arguments. just run it!
it works by creating additional debug pod and then attaching to it. 
Kubernetes with CRI is needed. Due to a technical limitation, squash-lite doesn't support 
scratch images at the moment (squash lite relys on the 'ls' command present in the image). 
`

func main() {
	var cfg kube.SquashConfig
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", descriptionUsage)
		flag.PrintDefaults()
	}

	flag.BoolVar(&cfg.NoClean, "no-clean", false, "don't clean temporar pod when existing")
	flag.BoolVar(&cfg.ChooseDebugger, "no-guess-debugger", false, "don't auto detect debugger to use")
	flag.BoolVar(&cfg.ChoosePod, "no-guess-pod", false, "don't auto detect pod to use")
	flag.IntVar(&cfg.TimeoutSeconds, "timeout", 60, "timeout in seconds to wait for debug pod to be ready")
	flag.Parse()

	err := kube.StartDebugContainer(cfg, nil)
	if err != nil {
		fmt.Println(err)
	}
}
