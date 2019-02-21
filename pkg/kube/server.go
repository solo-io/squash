package kube

import (
	"context"
	"fmt"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/debuggers"
	"github.com/solo-io/squash/pkg/debuggers/dlv"
	"github.com/solo-io/squash/pkg/debuggers/gdb"
	"github.com/solo-io/squash/pkg/debuggers/java"
	"github.com/solo-io/squash/pkg/debuggers/nodejs"
	"github.com/solo-io/squash/pkg/debuggers/python"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
)

func startDebugging(cfg Config, pid int) error {

	particularDebugger := getParticularDebugger(cfg.Attachment.Debugger)
	dbgServer, err := particularDebugger.Attach(pid)
	if err != nil {
		return err
	}

	if err := connectLocalPrepare(dbgServer, cfg.Attachment); err != nil {
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
	fmt.Println("proxy?")
	if dbgServer.Cmd() == nil {
		fmt.Println("proxy? - no")
		return nil
	}
	fmt.Println("proxy? - yes")
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

func connectLocalPrepare(dbgServer debuggers.DebugServer, att v1.DebugAttachment) error {
	// Some debuggers work best when connected "locally"
	// For these, we connect directly via `kubectl port-forward`
	// We write the target port to a CRD to be read from squashctl

	// get client
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		log.WithField("err", err).Error("getting debug attachment client")
		return err
	}

	// try to find a pre-existing CRD for this debug activity
	// create one if none exist
	da, err := findOrCreateDebugAttachmentCRD(ctx, daClient, att)
	if err != nil {
		return err
	}

	// set port value
	da.DebugServerAddress = fmt.Sprintf("inferfrompod:%v", dbgServer.Port())
	if _, err := (*daClient).Write(da, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {
		return err
	}

	return nil
}

func findOrCreateDebugAttachmentCRD(ctx context.Context, daClient *v1.DebugAttachmentClient, att v1.DebugAttachment) (*v1.DebugAttachment, error) {
	// don't need the error, just need to know if it exists
	da, _ := (*daClient).Read(att.Metadata.Namespace, att.Metadata.Name, clients.ReadOpts{Ctx: ctx})
	if da == nil {
		// need to create this debugAttachment
		newDa := &v1.DebugAttachment{
			Metadata: att.Metadata,
		}
		var err error
		da, err = (*daClient).Write(newDa, clients.WriteOpts{Ctx: ctx, OverwriteExisting: false})
		if err != nil {
			return nil, fmt.Errorf("Could not write debug attachment %v in namespace %v: %v", att.Metadata.Name, att.Metadata.Namespace, err)
		}
	}
	return da, nil
}

func startLocalServer() (net.Conn, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", ListenHost, options.OutPort))
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
