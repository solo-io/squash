package testutils

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/api/core/v1"
)

type E2eParams struct {
	DebugAttachmetName string

	kubectl *Kubectl
	Squash  *Squash

	Microservice1Pods       map[string]*v1.Pod
	Microservice2Pods       map[string]*v1.Pod
	CurrentMicroservicePod  *v1.Pod
	Current2MicroservicePod *v1.Pod
}

func NewE2eParams(daName string) E2eParams {
	k := NewKubectl("")
	return E2eParams{
		DebugAttachmetName: daName,

		kubectl: k,
		Squash:  NewSquash(k),

		Microservice1Pods: make(map[string]*v1.Pod),
		Microservice2Pods: make(map[string]*v1.Pod),
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
	fmt.Fprintf(GinkgoWriter, "creating environment %v \n", p.kubectl)

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
		case strings.HasPrefix(pod.ObjectMeta.Name, "squash-server"):
			panic(pod)
		case strings.HasPrefix(pod.ObjectMeta.Name, "squash-client"):
			panic(pod)
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

	p.kubectl.DeleteDebugAttachment(p.DebugAttachmetName)

	// wait for things to settle. may not be needed.
	time.Sleep(10 * time.Second)
}
func (p *E2eParams) Cleanup() {
	defer p.kubectl.StopProxy()
	defer p.kubectl.DeleteNS()
}
