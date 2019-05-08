package plank

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/debuggers/remote"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
)

func startDebugging(cfg *Config, pid int) error {

	particularDebugger := remote.GetParticularDebugger(cfg.Attachment.Debugger)
	dbgServer, err := particularDebugger.Attach(pid)
	if err != nil {
		return err
	}

	if err := connectLocalPrepare(cfg.ctx, dbgServer, cfg.Attachment); err != nil {
		return err
	}
	if err := proxyConnection(dbgServer); err != nil {
		return err
	}
	return nil
}

// we proxy so we can exit the debugger when disconnection occurs
// and so that we don't need to know the port the debugger is using
func proxyConnection(dbgServer remote.DebugServer) error {
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

func connectLocalPrepare(ctx context.Context, dbgServer remote.DebugServer, att v1.DebugAttachment) error {
	// Some debuggers work best when connected "locally"
	// For these, we connect directly via `kubectl port-forward`
	// We write the target port to a CRD to be read from squashctl

	// get client
	daClient, err := utils.GetBasicDebugAttachmentClient(ctx, plankKubeconfigPath)
	if err != nil {
		contextutils.LoggerFrom(ctx).With(zap.Error(err)).Error("getting debug attachment client")
		return err
	}

	// try to find a pre-existing CRD for this debug activity
	// create one if none exist
	da, err := daClient.Read(att.Metadata.Namespace, att.Metadata.Name, clients.ReadOpts{Ctx: ctx})
	if err != nil {
		return err
	}

	// set port value
	da.DebugServerAddress = fmt.Sprintf("inferfrompod:%v", dbgServer.Port())
	// write own plank pod name
	da.PlankName = os.Getenv("HOSTNAME")
	if _, err := daClient.Write(da, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {
		return err
	}

	return nil
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
