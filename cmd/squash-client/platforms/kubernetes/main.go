package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
	"github.com/solo-io/squash/pkg/debuggers/dlv"
	"github.com/solo-io/squash/pkg/debuggers/gdb"

	"github.com/solo-io/squash/pkg/platforms/kubernetes"
)

func main() {
	log.SetLevel(log.DebugLevel)

	customFormatter := new(log.TextFormatter)
	log.SetFormatter(customFormatter)

	log.Info("bridge started")

	err := debuggers.RunSquashClient(getDebugger, kubernetes.NewContainer2Pid())
	log.WithError(err).Fatal("Error running debug bridge")

}

func getDebugger(dbgtype string) debuggers.Debugger {
	var g gdb.GdbInterface
	var d dlv.DLV
	switch dbgtype {
	case "dlv":
		return &d
	case "gdb":
		return &g
	default:
		return nil
	}
}
