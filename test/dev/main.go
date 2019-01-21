package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/solo-io/squash/test/devutil"
	"github.com/solo-io/squash/test/devutil/writer"
)

type Cfg struct {
	init      bool
	att       bool
	clean     bool
	name      string
	namespace string
}

func main() {
	cfg := Cfg{}
	flag.BoolVar(&cfg.init, "init", false, "create base resources")
	flag.BoolVar(&cfg.att, "att", false, "attach a debugger")
	flag.BoolVar(&cfg.clean, "clean", false, "remove debugger")
	flag.StringVar(&cfg.name, "name", "mitch", "resource name")
	flag.StringVar(&cfg.namespace, "namespace", "squash-debugger-test", "resource name")
	flag.Parse()

	if err := run(cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cfg Cfg) error {
	kWriter := writer.New(os.Stdout)
	params, err := devutil.NewE2eParams(cfg.namespace, cfg.name, kWriter)
	if err != nil {
		return err
		// return params.Cleanup()
	}

	if cfg.init {
		if err := params.SetupDev(); err != nil {
			return params.Cleanup()
		}
	}
	if cfg.init && cfg.att {
		container := params.CurrentMicroservicePod.Spec.Containers[0]
		dbgattachment, err := params.UserController.Attach(cfg.name, params.Namespace, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
		if err != nil {
			return err
		}
		fmt.Println(dbgattachment)
	}
	if cfg.clean {
		return params.Cleanup()
	}

	return nil

}
