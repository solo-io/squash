package testutils

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/squash/pkg/actions"
	"github.com/solo-io/squash/test/testutils/kubecdl"

	"k8s.io/api/core/v1"
)

type E2eParams struct {
	DebugAttachmetName string

	Namespace      string
	kubectl        *kubecdl.Kubecdl
	Squash         *Squash
	UserController actions.UserController

	AgentPod                map[string]*v1.Pod
	Microservice1Pods       map[string]*v1.Pod
	Microservice2Pods       map[string]*v1.Pod
	CurrentMicroservicePod  *v1.Pod
	Current2MicroservicePod *v1.Pod

	crbAdminName string
}

func NewE2eParams(namespace, daName string, w io.Writer) E2eParams {
	k := kubecdl.NewKubecdl(namespace, "", w)
	uc, err := actions.NewUserController()
	Expect(err).NotTo(HaveOccurred())

	return E2eParams{
		DebugAttachmetName: daName,

		Namespace:      k.Namespace,
		kubectl:        k,
		Squash:         NewSquash(k),
		UserController: uc,

		AgentPod:          make(map[string]*v1.Pod),
		Microservice1Pods: make(map[string]*v1.Pod),
		Microservice2Pods: make(map[string]*v1.Pod),

		crbAdminName: "serviceaccount-cluster-admin-level",
	}
}

func (p *E2eParams) SetupE2e() {

	if err := p.kubectl.Proxy(); err != nil {
		fmt.Fprintln(GinkgoWriter, "error creating ns", err)
		panic(err)
	}

	if err := p.kubectl.CreateNS(); err != nil {
		fmt.Fprintln(GinkgoWriter, "error creating ns", err)
		panic(err)
	}
	// give the namespace time to be created
	time.Sleep(time.Second)

	fmt.Fprintf(GinkgoWriter, "creating environment %v \n", p.kubectl)

	if err := p.kubectl.CreateSleep("../../contrib/kubernetes/squash-agent.yaml"); err != nil {
		panic(err)
	}
	if err := p.kubectl.Create("../../contrib/example/service1/service1.yml"); err != nil {
		panic(err)
	}
	if err := p.kubectl.Create("../../contrib/example/service2/service2.yml"); err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	err := p.kubectl.WaitPods(ctx)
	cancel()
	Expect(err).NotTo(HaveOccurred())

	pods, err := p.kubectl.Pods()
	Expect(err).NotTo(HaveOccurred())

	for _, pod := range pods.Items {
		// make a copy
		newpod := pod
		switch {
		case strings.HasPrefix(pod.ObjectMeta.Name, "example-service1"):
			p.Microservice1Pods[pod.Spec.NodeName] = &newpod
		case strings.HasPrefix(pod.ObjectMeta.Name, "example-service2"):
			p.Microservice2Pods[pod.Spec.NodeName] = &newpod
		case strings.HasPrefix(pod.ObjectMeta.Name, "squash-agent"):
			pathToClientBinary := "../../target/agent/squash-agent"
			if _, err := os.Stat(pathToClientBinary); os.IsNotExist(err) {
				Fail("You must generate the squash-agent binary before running this e2e test.")
			}
			// replace squash server and client binaries with local binaries for easy debuggings
			Must(p.kubectl.Cp(pathToClientBinary, "/tmp/", pod.ObjectMeta.Name, "squash-agent"))

			// client is in host pid namespace, so can't write logs to pid 1. use the fact that the client has the pod name in the env.
			clientscript := "SLEEPPID=$(for pid in $(pgrep sleep); do if grep --silent " + pod.ObjectMeta.Name + " /proc/$pid/environ; then echo $pid;fi; done) && "
			clientscript += " /tmp/squash-agent  > /proc/$SLEEPPID/fd/1 2> /proc/$SLEEPPID/fd/2"
			Must(p.kubectl.ExecAsync(pod.ObjectMeta.Name, "squash-agent", "sh", "-c", clientscript))
			p.AgentPod[pod.Spec.NodeName] = &newpod
		}
	}

	// choose one of the microservice pods to be our victim.
	for _, v := range p.Microservice1Pods {
		p.CurrentMicroservicePod = v
		break
	}
	if p.CurrentMicroservicePod == nil {
		Fail("can't find service pod")
	}
	for _, v := range p.Microservice2Pods {
		p.Current2MicroservicePod = v
		break
	}
	if p.CurrentMicroservicePod == nil {
		Fail("can't find service2 pod")
	}

	if len(p.AgentPod) == 0 {
		Fail("can't find client pods")
	}

	if p.AgentPod[p.CurrentMicroservicePod.Spec.NodeName] == nil {
		Fail("can't find client pods")
	}

	if err := p.kubectl.GrantClusterAdminPermissions(p.crbAdminName); err != nil {
		Fail(fmt.Sprintf("Failed to create permissions: %v", err))
	}

	p.kubectl.DeleteDebugAttachment(p.DebugAttachmetName)

	// wait for things to settle. may not be needed.
	time.Sleep(10 * time.Second)
}

func (p *E2eParams) CleanupE2e() {
	defer p.kubectl.StopProxy()
	// Deleting namespaces can be slow, do it in the background
	defer func() {
		go p.kubectl.DeleteNS()
		// give kubectl syscall time to execute
		time.Sleep(100 * time.Millisecond)
	}()

	if err := p.kubectl.RemoveClusterAdminPermissions(p.crbAdminName); err != nil {
		// No need to fail on these errors
		fmt.Sprintf("Failed to delete permissions: %v", err)
	}

	clogs, _ := p.kubectl.Logs(p.AgentPod[p.CurrentMicroservicePod.Spec.NodeName].ObjectMeta.Name)
	fmt.Fprintln(GinkgoWriter, "client logs:")
	fmt.Fprintln(GinkgoWriter, string(clogs))
}

func Must(err error) {
	Expect(err).NotTo(HaveOccurred())
}
