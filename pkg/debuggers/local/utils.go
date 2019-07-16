package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
)

func GetPortForwardCmd(targetName, targetNamespace string, localPort, targetRemotePort int) *exec.Cmd {
	portSpec := fmt.Sprintf("%v:%v", localPort, targetRemotePort)
	cmd := exec.Command("kubectl", "port-forward", targetName, portSpec, "-n", targetNamespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd
}

func getPortForwardWithRandomLocalCmd(targetName, targetNamespace string, targetRemotePort int) string {
	portSpec := fmt.Sprintf(":%v", targetRemotePort)
	return strings.Join([]string{"kubectl", "port-forward", targetName, portSpec, "-n", targetNamespace}, " ")
}

func GetParticularDebugger(dbgtype string) Local {
	var g GdbInterface
	var d DLV
	var j JavaInterface
	var jp JavaPortInterface
	var p PythonInterface
	var n NodeJsDebugger

	switch dbgtype {
	case "dlv":
		return &d
	case "gdb":
		return &g
	case "java":
		return &j
	case "java-port":
		return &jp
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

func GetDebugPortFromCrd(ctx context.Context, daClient v1.DebugAttachmentClient, daName, daNamespace string) (int, error) {
	// TODO - all of our ports should be gotten from the crd. As is, it is possible that the random port chosen from ip_addr:0 could return 1236 - slim chance but may as well handle it
	da, err := waitForDebugServerAddress(ctx, daClient, daName, daNamespace)
	if err != nil {
		return 0, fmt.Errorf("Could not read debug attachment %v in namespace %v: %v", daName, daNamespace, err)
	}
	port, err := da.GetPortFromDebugServerAddress()
	if err != nil {
		return 0, err
	}
	return port, nil
}

func waitForDebugServerAddress(ctx context.Context, daClient v1.DebugAttachmentClient, daName, daNamespace string) (*v1.DebugAttachment, error) {
	dac, errc, err := daClient.Watch(daNamespace, clients.WatchOpts{Ctx: ctx})
	if err != nil {
		return &v1.DebugAttachment{}, err
	}

	// TODO - make timeout configurable
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()
	for {
		select {
		case err, _ := <-errc:
			return &v1.DebugAttachment{}, err
		case <-ctx.Done():
			return &v1.DebugAttachment{}, fmt.Errorf("could not find debug spec in the allotted time.")
		case das, ok := <-dac:
			if !ok {
				return &v1.DebugAttachment{}, fmt.Errorf("could not read watch channel")
			}

			if len(das) == 0 {
				continue
			}

			da := checkDebugAttachmentsForAddress(das, daName)
			if da != nil {
				return da, nil
			}
		}
	}
}

func checkDebugAttachmentsForAddress(das v1.DebugAttachmentList, daName string) *v1.DebugAttachment {
	for _, da := range das {
		if da.Metadata.Name == daName && da.DebugServerAddress != "" {
			return da
		}
	}
	return nil
}
