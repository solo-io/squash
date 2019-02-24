package local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
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
	da, err := waitForDebugServerAddress(daName, daNamespace)
	if err != nil {
		return 0, fmt.Errorf("Could not read debug attachment %v in namespace %v: %v", daName, daNamespace, err)
	}
	port, err := da.GetPortFromDebugServerAddress()
	if err != nil {
		return 0, err
	}
	return port, nil
}

func waitForDebugServerAddress(daName, daNamespace string) (*v1.DebugAttachment, error) {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		log.WithField("err", err).Error("getting debug attachment client")
		return &v1.DebugAttachment{}, err
	}
	dac, errc, err := (*daClient).Watch(daNamespace, clients.WatchOpts{Ctx: ctx})
	if err != nil {
		return &v1.DebugAttachment{}, nil
	}
	var cancel context.CancelFunc = func() {}
	defer cancel()
	for {
		select {
		case err, _ := <-errc:
			return &v1.DebugAttachment{}, err
		case <-time.After(10 * time.Second):
			// TODO - make timeout configurable, better error message
			return &v1.DebugAttachment{}, fmt.Errorf("Could not find debug spec in the allotted time.")
		case das, ok := <-dac:
			if !ok {
				return &v1.DebugAttachment{}, fmt.Errorf("could not read watch channel")
			}
			cancel()

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
	return &v1.DebugAttachment{}
}
