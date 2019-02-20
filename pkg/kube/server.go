package kube

import (
	"fmt"
	"io"
	"net"

	"github.com/solo-io/squash/pkg/debuggers"
	"github.com/solo-io/squash/pkg/debuggers/dlv"
	"github.com/solo-io/squash/pkg/debuggers/gdb"
	"github.com/solo-io/squash/pkg/debuggers/java"
	"github.com/solo-io/squash/pkg/debuggers/nodejs"
	"github.com/solo-io/squash/pkg/debuggers/python"
)

func startDebugging(cfg Config, pid int) error {

	particularDebugger := getParticularDebugger(cfg.Attachment.Debugger)
	dbgServer, err := particularDebugger.Attach(pid)
	if err != nil {
		return err
	}

	if err := connectLocalPrepare(dbgServer); err != nil {
		return err
	}
	if err := proxyConnection(dbgServer); err != nil {
		return err
	}
	return nil
}

// we proxy so we can exit the debugger when disconnection occurs
// and so that we don't need to know the port the debugger is using
func proxyConnection(dbgServer debuggers.DebugServer) error {
	// only proxy the debuggers that are called by this process
	if dbgServer.Cmd() == nil {
		return nil
	}
	errchan := make(chan error, 1)
	reporterr := func(err error) {
		select {
		case errchan <- err:
		default:
		}
	}
	go func() {
		reporterr(dbgServer.Cmd().Wait())
	}()

	conn, err := startLocalServer()
	if err != nil {
		return err
	}
	defer conn.Close()

	// connect to debug server
	conn2, err := net.Dial("tcp", fmt.Sprintf("%v:%v", ListenHost, dbgServer.Port()))
	if err != nil {
		return err
	}

	go func() {
		_, err := io.Copy(conn2, conn)
		reporterr(err)
	}()
	go func() {
		// if the client ends the session - no error
		io.Copy(conn, conn2)
		reporterr(nil)
	}()

	return <-errchan
}

// TODO
func connectLocalPrepare(dbgServer debuggers.DebugServer) error {
	// Some debuggers work best when connected "locally"
	// For these, we connect directly via `kubectl port-forward`
	// We write the target port to a CRD to be read from squashctl

	// get client

	// try to find a pre-existing CRD for this debug activity
	// create one if none exist
	// findOrCreateDebugAttachmentCRD

	// set port value

	return nil
}

func startLocalServer() (net.Conn, error) {
	l, err := net.Listen("tcp", ListenHost+":"+OutPort)
	if err != nil {
		return nil, err
	}
	defer l.Close()
	conn, err := l.Accept()
	return conn, err

}

func getParticularDebugger(dbgtype string) debuggers.Debugger {
	var g gdb.GdbInterface
	var d dlv.DLV
	var j java.JavaInterface
	var p python.PythonInterface

	switch dbgtype {
	case "dlv":
		return &d
	case "gdb":
		return &g
	case "java":
		return &j
	case "nodejs":
		return nodejs.NewNodeDebugger(nodejs.DebuggerPort)
	case "nodejs8":
		return nodejs.NewNodeDebugger(nodejs.InspectorPort)
	case "python":
		return &p
	default:
		return nil
	}
}
