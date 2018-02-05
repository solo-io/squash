package e2e_test

import (
	"context"
	"fmt"
	"strings"
	"time"

	//	. "github.com/solo-io/squash/test/e2e"
	//"k8s.io/client-go/pkg/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/squash/test/utils"
	"k8s.io/api/core/v1"
)

//const KubeEndpoint = "http://localhost:8001/api"

func Must(err error) {
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("Single debug mode", func() {

	var (
		kubectl *Kubectl
		squash  *Squash

		ServerPod              *v1.Pod
		ClientPods             map[string]*v1.Pod
		MicroservicePods       map[string]*v1.Pod
		CurrentMicroservicePod *v1.Pod
	)

	BeforeEach(func() {
		ServerPod = nil
		ClientPods = make(map[string]*v1.Pod)
		MicroservicePods = make(map[string]*v1.Pod)
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

		if err := kubectl.CreateLocalRolesAndSleep("../../contrib/kubernetes/squash-server.yml"); err != nil {
			panic(err)
		}
		if err := kubectl.CreateSleep("../../contrib/kubernetes/squash-client.yml"); err != nil {
			panic(err)
		}
		if err := kubectl.Create("../../contrib/example/service1/service1.yml"); err != nil {
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
				MicroservicePods[pod.Spec.NodeName] = &newpod
			case strings.HasPrefix(pod.ObjectMeta.Name, "squash-server"):
				// replace squash server and client binaries with local binaries for easy debuggings
				Must(kubectl.Cp("../../target/squash-server/squash-server", "/tmp/", pod.ObjectMeta.Name, "squash-server"))
				Must(kubectl.ExecAsync(pod.ObjectMeta.Name, "squash-server", "sh", "-c", "/tmp/squash-server --cluster=kube --host=0.0.0.0 --port=8080 > /proc/1/fd/1 2> /proc/1/fd/2"))
				ServerPod = &newpod
			case strings.HasPrefix(pod.ObjectMeta.Name, "squash-client"):
				// replace squash server and client binaries with local binaries for easy debuggings
				Must(kubectl.Cp("../../target/squash-client/squash-client", "/tmp/", pod.ObjectMeta.Name, "squash-client"))

				// client is in host pid namespace, so can't write logs to pid 1. use the fact that the client has the pod name in the env.
				clientscript := "SLEEPPID=$(for pid in $(pgrep sleep); do if grep --silent " + pod.ObjectMeta.Name + " /proc/$pid/environ; then echo $pid;fi; done) && "
				clientscript += " /tmp/squash-client  > /proc/$SLEEPPID/fd/1 2> /proc/$SLEEPPID/fd/2"
				Must(kubectl.ExecAsync(pod.ObjectMeta.Name, "squash-client", "sh", "-c", clientscript))
				ClientPods[pod.Spec.NodeName] = &newpod
			}
		}

		// choose one of the microservice pods to be our victim.
		for _, v := range MicroservicePods {
			CurrentMicroservicePod = v
			break
		}

		// wait for things to settle. may not be needed.
		time.Sleep(10 * time.Second)
	})

	AfterEach(func() {

		logs, _ := kubectl.Logs(ServerPod.ObjectMeta.Name)
		fmt.Fprintln(GinkgoWriter, "server logs:")
		fmt.Fprintln(GinkgoWriter, string(logs))
		clogs, _ := kubectl.Logs(ClientPods[CurrentMicroservicePod.Spec.NodeName].ObjectMeta.Name)
		fmt.Fprintln(GinkgoWriter, "client logs:")
		fmt.Fprintln(GinkgoWriter, string(clogs))

		//		fmt.Println("ZBAM", ClientPods[CurrentMicroservicePod.Spec.NodeName].ObjectMeta.Name, len(clogs))
		//		time.Sleep(2 * time.Minute)

		kubectl.DeleteNS()
		kubectl.StopProxy()
	})

	Describe("Single Container mode", func() {
		It("should get a debug server endpoint", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			updatedattachment, err := squash.Wait(dbgattachment.Metadata.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.DebugServerAddress).ToNot(BeEmpty())
		})

		It("should get a debug server endpoint, specific process", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "service1", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)

			updatedattachment, err := squash.Wait(dbgattachment.Metadata.Name)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.DebugServerAddress).ToNot(BeEmpty())
		})

		It("should get a debug server endpoint, specific process that doesn't exist", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "processNameDoesntExist", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			updatedattachment, err := squash.Wait(dbgattachment.Metadata.Name)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).To(Equal("error"))
		})

		It("Be able to re-attach once session exited", func() {
			container := CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := squash.Attach(container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			updatedattachment, err := squash.Wait(dbgattachment.Metadata.Name)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.DebugServerAddress).ToNot(BeEmpty())

			// Ok; now delete the attachment
			err = squash.Delete(dbgattachment)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)

			// try again!
			dbgattachment, err = squash.Attach(container.Image, CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			updatedattachment, err = squash.Wait(dbgattachment.Metadata.Name)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.DebugServerAddress).ToNot(BeEmpty())

			// kube portfowrad...
			// proc, localaddr, err := kubectl.PortForward(updatedattachment.Status.DebugServerAddress)
			// Expect(err).NotTo(HaveOccurred())
			// defer proc.Kill()
			// fmt.Println("Connect " + localaddr)
			// err = exec.Command("sleep", "1h").Run()
			// Expect(err).NotTo(HaveOccurred())
			// dlv connect localaddr
			// dlv exit
			//			exec.Cmd()

			// connect with dlv and detach
			// dlv connect updatedattachment.Status.DebugServerAddress
		})
	})

	/*
		Describe("Service mode", func() {
			It("should get a debug server endpoint", func() {

				var p *v1.Pod
				for _, v := range MicroservicePods {
					p = v
					break
				}

				container := p.Spec.Containers[0]
				dbgconfig, err := squash.Service(container.Image, "example-service1-svc", "dlv", "main.go:80")
				if err != nil {
					logs, _ := kubectl.Logs(ServerPod.ObjectMeta.Name)
					fmt.Println(string(logs))
					panic(err)
				}
				time.Sleep(60 * time.Second)

				serviceurl := fmt.Sprintf(KubeEndpoint+"/v1/namespaces/%s/services/example-service1-svc/proxy/calc", kubectl.Namespace)
				go func() {
					_, err := http.Get(serviceurl)
					if err != nil {
						fmt.Println("Error issuing http request ", err)
					} else {
						fmt.Println("http request OK")

					}

				}()
				dbgsession, err := squash.Wait(dbgconfig.ID)
				if err != nil {
					logs, _ := kubectl.Logs(ServerPod.ObjectMeta.Name)
					fmt.Println("server logs:")
					fmt.Println(string(logs))
					clogs, _ := kubectl.Logs(ClientPods[p.Spec.NodeName].ObjectMeta.Name)
					fmt.Println("client logs:")
					fmt.Println(string(clogs))
					panic(err)
				}

				Expect(dbgsession).ToNot(Equal(nil))

			})
		})
	*/
})
