package devutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/solo-io/squash/pkg/actions"
	"github.com/solo-io/squash/test/testutils"
	"github.com/solo-io/squash/test/testutils/kubecdl"
	"k8s.io/api/core/v1"
)

type E2eParams struct {
	DebugAttachmetName string

	Namespace      string
	kubectl        *kubecdl.Kubecdl
	Squash         *testutils.Squash
	UserController actions.UserController

	ClientPods              map[string]*v1.Pod
	Microservice1Pods       map[string]*v1.Pod
	Microservice2Pods       map[string]*v1.Pod
	CurrentMicroservicePod  *v1.Pod
	Current2MicroservicePod *v1.Pod

	crbAdminName string
}

func NewE2eParams(namespace, daName string, w io.Writer) (E2eParams, error) {
	k := kubecdl.NewKubecdl(namespace, "", w)
	uc, err := actions.NewUserController()
	if err != nil {
		return E2eParams{}, err
	}

	return E2eParams{
		DebugAttachmetName: daName,

		Namespace:      k.Namespace,
		kubectl:        k,
		Squash:         testutils.NewSquash(k),
		UserController: uc,

		ClientPods:        make(map[string]*v1.Pod),
		Microservice1Pods: make(map[string]*v1.Pod),
		Microservice2Pods: make(map[string]*v1.Pod),

		crbAdminName: "serviceaccount-cluster-admin-level",
	}, nil
}
func (p *E2eParams) SetupDev() error {

	if err := p.kubectl.Proxy(); err != nil {
		fmt.Println("error creating ns", err)
		return err
	}

	if err := p.kubectl.CreateNS(); err != nil {
		fmt.Println("error creating ns", err)
		return err
	}
	fmt.Printf("creating environment %v \n", p.kubectl)

	if err := p.kubectl.CreateSleep("../../contrib/kubernetes/squash-client.yml"); err != nil {
		return err
	}
	if err := p.kubectl.Create("../../contrib/example/service1/service1.yml"); err != nil {
		return err
	}
	if err := p.kubectl.Create("../../contrib/example/service2/service2.yml"); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	err := p.kubectl.WaitPods(ctx)
	cancel()
	if err != nil {
		return err
	}

	pods, err := p.kubectl.Pods()
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		// make a copy
		newpod := pod
		switch {
		case strings.HasPrefix(pod.ObjectMeta.Name, "example-service1"):
			p.Microservice1Pods[pod.Spec.NodeName] = &newpod
		case strings.HasPrefix(pod.ObjectMeta.Name, "example-service2"):
			p.Microservice2Pods[pod.Spec.NodeName] = &newpod
		case strings.HasPrefix(pod.ObjectMeta.Name, "squash-server"):
			return err
		case strings.HasPrefix(pod.ObjectMeta.Name, "squash-client"):
			pathToClientBinary := "../../target/squash-client/squash-client"
			if _, err := os.Stat(pathToClientBinary); os.IsNotExist(err) {
				return fmt.Errorf("You must generate the squash-client binary before running this e2e test.")
			}
			// replace squash server and client binaries with local binaries for easy debuggings
			err := p.kubectl.Cp(pathToClientBinary, "/tmp/", pod.ObjectMeta.Name, "squash-client")
			if err != nil {
				return err
			}

			// client is in host pid namespace, so can't write logs to pid 1. use the fact that the client has the pod name in the env.
			clientscript := "SLEEPPID=$(for pid in $(pgrep sleep); do if grep --silent " + pod.ObjectMeta.Name + " /proc/$pid/environ; then echo $pid;fi; done) && "
			clientscript += " /tmp/squash-client  > /proc/$SLEEPPID/fd/1 2> /proc/$SLEEPPID/fd/2"
			err = p.kubectl.ExecAsync(pod.ObjectMeta.Name, "squash-client", "sh", "-c", clientscript)
			if err != nil {
				return err
			}
			p.ClientPods[pod.Spec.NodeName] = &newpod
		}
	}

	// choose one of the microservice pods to be our victim.
	for _, v := range p.Microservice1Pods {
		p.CurrentMicroservicePod = v
		break
	}
	if p.CurrentMicroservicePod == nil {
		return fmt.Errorf("can't find service pod")
	}
	for _, v := range p.Microservice2Pods {
		p.Current2MicroservicePod = v
		break
	}
	if p.CurrentMicroservicePod == nil {
		return fmt.Errorf("can't find service2 pod")
	}

	if len(p.ClientPods) == 0 {
		return fmt.Errorf("can't find client pods")
	}

	if p.ClientPods[p.CurrentMicroservicePod.Spec.NodeName] == nil {
		return fmt.Errorf("can't find client pods")
	}

	if err := p.kubectl.GrantClusterAdminPermissions(p.crbAdminName); err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to create permissions: %v", err))
	}

	p.kubectl.DeleteDebugAttachment(p.DebugAttachmetName)

	// wait for things to settle. may not be needed.
	time.Sleep(10 * time.Second)
	p.PrintLogs()
	return nil
}

func (p *E2eParams) Cleanup() error {
	defer p.kubectl.StopProxy()
	defer p.kubectl.DeleteNS()

	if err := p.kubectl.RemoveClusterAdminPermissions(p.crbAdminName); err != nil {
		return fmt.Errorf("Failed to delete permissions: %v", err)
	}

	return nil
}

func (p *E2eParams) PrintLogs() error {
	clogs, err := p.kubectl.Logs(p.ClientPods[p.CurrentMicroservicePod.Spec.NodeName].ObjectMeta.Name)
	if err != nil {
		return err
	}
	fmt.Println("client logs:")
	fmt.Println(string(clogs))
	return nil
}
