package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	squashcli "github.com/solo-io/squash/pkg/cmd/cli"
	. "github.com/solo-io/squash/test/utils"
	"k8s.io/api/core/v1"
)

//const KubeEndpoint = "http://localhost:8001/api"

func Must(err error) {
	Expect(err).NotTo(HaveOccurred())
}

var tmpDaName = "mitch"

var _ = Describe("Single debug mode", func() {

	var (
		kubectl *Kubectl
		squash  *Squash

		ServerPod               *v1.Pod
		ClientPods              map[string]*v1.Pod
		Microservice1Pods       map[string]*v1.Pod
		Microservice2Pods       map[string]*v1.Pod
		CurrentMicroservicePod  *v1.Pod
		Current2MicroservicePod *v1.Pod
	)

	/*

		Deploy the services that you will debug

	*/
	BeforeEach(func() {
		ServerPod = nil
		ClientPods = make(map[string]*v1.Pod)
		Microservice1Pods = make(map[string]*v1.Pod)
		Microservice2Pods = make(map[string]*v1.Pod)
		kubectl = NewKubectl("")

		if err := kubectl.Proxy(); err != nil {
			fmt.Fprintln(GinkgoWriter, "error creating ns", err)
			panic(err)
		}

		if err := kubectl.CreateNS(); err != nil {
			fmt.Fprintln(GinkgoWriter, "error creating ns", err)
			panic(err)
		}
		fmt.Fprintf(GinkgoWriter, "creating environment %v \n", kubectl)

		if err := kubectl.Create("../../contrib/example/service1/service1.yml"); err != nil {
			panic(err)
		}
		if err := kubectl.Create("../../contrib/example/service2/service2.yml"); err != nil {
			panic(err)
		}

		squash = NewSquash(kubectl)
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		err := kubectl.WaitPods(ctx)
		cancel()
		Expect(err).NotTo(HaveOccurred())

		pods, err := kubectl.Pods()
		Expect(err).NotTo(HaveOccurred())

		for _, pod := range pods.Items {
			// make a copy
			newpod := pod
			switch {
			case strings.HasPrefix(pod.ObjectMeta.Name, "example-service1"):
				Microservice1Pods[pod.Spec.NodeName] = &newpod
			case strings.HasPrefix(pod.ObjectMeta.Name, "example-service2"):
				Microservice2Pods[pod.Spec.NodeName] = &newpod
			case strings.HasPrefix(pod.ObjectMeta.Name, "squash-server"):
				panic(pod)
			case strings.HasPrefix(pod.ObjectMeta.Name, "squash-client"):
				panic(pod)
			}
		}

		// choose one of the microservice pods to be our victim.
		for _, v := range Microservice1Pods {
			CurrentMicroservicePod = v
			break
		}
		if CurrentMicroservicePod == nil {
			Fail("can't find service pod")
		}
		for _, v := range Microservice2Pods {
			Current2MicroservicePod = v
			break
		}
		if CurrentMicroservicePod == nil {
			Fail("can't find service2 pod")
		}

		kubectl.DeleteDebugAttachment(tmpDaName)

		// wait for things to settle. may not be needed.
		time.Sleep(10 * time.Second)
	})

	AfterEach(func() {
		defer kubectl.StopProxy()
		defer kubectl.DeleteNS()

		logs, _ := kubectl.Logs(ServerPod.ObjectMeta.Name)
		fmt.Fprintln(GinkgoWriter, "server logs:")
		fmt.Fprintln(GinkgoWriter, string(logs))
		clogs, _ := kubectl.Logs(ClientPods[CurrentMicroservicePod.Spec.NodeName].ObjectMeta.Name)
		fmt.Fprintln(GinkgoWriter, "client logs:")
		fmt.Fprintln(GinkgoWriter, string(clogs))
	})

	Describe("Single Container mode", func() {
		It("should get a debug server endpoint", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(tmpDaName, container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})

		It("should get a debug server endpoint, specific process", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(tmpDaName, container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "service1", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)

			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})

		It("should get a debug server endpoint, specific process that doesn't exist", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(tmpDaName, container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "processNameDoesntExist", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).NotTo(Equal(core.Status_Accepted))
		})
		FIt("should attach to two micro services", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(tmpDaName, container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).To(Equal(core.Status_Accepted))

			container = Current2MicroservicePod.Spec.Containers[0]
			dbgattachment, err = squash.Attach(tmpDaName, container.Image, Current2MicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			updatedattachment, err = squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).To(Equal(core.Status_Accepted))
		})

		It("Be able to re-attach once session exited", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(tmpDaName, container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())

			// Ok; now delete the attachment
			err = squash.Delete(dbgattachment)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)

			// try again!
			dbgattachment, err = squash.Attach(tmpDaName, container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			updatedattachment, err = squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})
	})

})
