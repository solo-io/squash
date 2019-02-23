package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/utils"
)

func GetPortForwardCmd(targetName, targetNamespace string, localPort, targetRemotePort int) *exec.Cmd {
	portSpec := fmt.Sprintf("%v:%v", localPort, targetRemotePort)
	cmd := exec.Command("kubectl", "port-forward", targetName, portSpec, "-n", targetNamespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func GetParticularDebugger(dbgtype string) Local {
	var g GdbInterface
	var d DLV
	var j JavaInterface
	var p PythonInterface
	var n NodeJsDebugger

	switch dbgtype {
	case "dlv":
		return &d
	case "gdb":
		return &g
	case "java":
		return &j
	case "nodejs":
		return &n
	case "nodejs8":
		return &n
	case "python":
		return &p
	default:
		return nil
	}
}

func GetDebugPortFromCrd(daName, daNamespace string) (int, error) {
	// TODO - all of our ports should be gotten from the crd. As is, it is possible that the random port chosen from ip_addr:0 could return 1236 - slim chance but may as well handle it
	// Give debug container time to create the CRD
	// TODO - reduce this sleep time
	time.Sleep(5 * time.Second)
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		log.WithField("err", err).Error("getting debug attachment client")
		return 0, err
	}
	da, err := (*daClient).Read(daNamespace, daName, clients.ReadOpts{Ctx: ctx})
	if err != nil {
		return 0, fmt.Errorf("Could not read debug attachment %v in namespace %v: %v", daName, daNamespace, err)
	}
	port, err := da.GetPortFromDebugServerAddress()
	if err != nil {
		return 0, err
	}
	return port, nil
}
