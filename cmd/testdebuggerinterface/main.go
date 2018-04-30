package main

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/squash/pkg/debuggers"
	"github.com/solo-io/squash/pkg/debuggers/dlv"
	"github.com/solo-io/squash/pkg/debuggers/gdb"
)

func main() {
	log.SetLevel(log.DebugLevel)

	customFormatter := new(log.TextFormatter)
	log.SetFormatter(customFormatter)

	log.Info("bridge started")
	pid, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	if true {
		ds, err := getDebugger(os.Args[1]).StartDebugServer(pid)
		if err != nil {
			panic(err)
		}
		fmt.Println(ds.Port())

	} else {

		l, err := getDebugger(os.Args[1]).AttachTo(pid)

		if err != nil {
			panic(err)
		}
		err = l.SetBreakpoint("main.go:80")
		if err != nil {
			panic(err)
		}
		e, err := l.Continue()

		if err != nil {
			panic(err)
		}
		<-e
		ds, err := l.IntoDebugServer()
		if err != nil {
			panic(err)
		}

		fmt.Println(ds.Port())
	}
	select {}

	// getDebugger(os.Args[1]).StartDebugServer(pid)
}

func getDebugger(dbgtype string) debuggers.Debugger {
	var g gdb.GdbInterface
	var d dlv.DLV
	switch dbgtype {
	case "dlv":
		return &d
	case "gdb":
		fallthrough
	default:
		return &g
	}
}
