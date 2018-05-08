package main

import (
	"flag"
	"fmt"

	"github.com/solo-io/squash/pkg/lite/kube"
)

func main() {
	var cfg kube.SquashConfig
	flag.BoolVar(&cfg.NoClean, "no-clean", false, "don't clean temporar pod when existing")
	flag.BoolVar(&cfg.ChooseDebugger, "no-guess-debugger", false, "don't auto detect debugger to use")
	flag.BoolVar(&cfg.ChoosePod, "no-guess-pod", false, "don't auto detect pod to use")
	flag.IntVar(&cfg.TimeoutSeconds, "timeout", 60, "timeout in seconds to wait for debug pod to be ready")
	flag.Parse()

	err := kube.StartDebugContainer(cfg)
	if err != nil {
		fmt.Println(err)
	}
}
