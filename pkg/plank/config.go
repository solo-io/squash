package plank

import (
	"context"
	"fmt"
	"os"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	sqOpts "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	Attachment v1.DebugAttachment
	// This is the debugger installed in the container
	// Options are dlv or gdb
	Debugger   string
	kubeClient *kubernetes.Clientset
	daClient   *v1.DebugAttachmentClient
	ctx        context.Context
}

func GetConfig(ctx context.Context) (*Config, error) {
	debugNamespace := os.Getenv(sqOpts.PlankEnvDebugAttachmentNamespace)
	daName := os.Getenv(sqOpts.PlankEnvDebugAttachmentName)
	plankName := os.Getenv(sqOpts.KubeEnvPodName)

	restCfg, err := gokubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return nil, err
	}

	// a debug attachment should have been created, pick it up
	da, err := (*daClient).Read(debugNamespace, daName, clients.ReadOpts{})
	if err != nil {
		return nil, err
	}
	if err := validateDebugAttachmentForPlankInit(da); err != nil {
		return nil, err
	}
	da.PlankName = plankName
	da, err = (*daClient).Write(da, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true})
	if err != nil {
		return nil, err
	}

	return &Config{
		kubeClient: kubeClient,
		daClient:   daClient,
		Attachment: *da,
		Debugger:   os.Getenv(sqOpts.PlankDockerEnvDebuggerType),
		ctx:        ctx,
	}, nil
}

// assert all the requirements for a debug attachment when it is read by a plank pod during startup
func validateDebugAttachmentForPlankInit(da *v1.DebugAttachment) error {
	errorMsg := ""
	assertNotNilString(&errorMsg, da.Pod, "Pod")
	assertNotNilString(&errorMsg, da.Container, "Container")
	assertNotNilString(&errorMsg, da.Debugger, "Debugger")
	assertNotNilString(&errorMsg, da.DebugNamespace, "DebugNamespace")
	if errorMsg != "" {
		return fmt.Errorf("Invalid Debug Attachment for Plank init: %v", errorMsg)
	}
	return nil
}

func assertNotNilString(errs *string, value, name string) {
	if value == "" {
		*errs = fmt.Sprintf("%v\n field %v should not be empty", errs, name)
	}
}
