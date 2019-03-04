package e2e_test

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/squash/pkg/api/v1"
	sqOpts "github.com/solo-io/squash/pkg/options"
	squashcli "github.com/solo-io/squash/pkg/squashctl"
	"github.com/solo-io/squash/test/testutils"
)

//const KubeEndpoint = "http://localhost:8001/api"

func Must(err error) {
	Expect(err).NotTo(HaveOccurred())
}

var (
	daName  = "debug-attachment-1"
	daName2 = "debug-attachment-2"
	// testNamespace = "squash-debugger-test2"
	testNamespace = "stest"
	testNSRoot    = "stest"
	testsStarted  = 0
)

var _ = Describe("Single debug mode", func() {

	seed := time.Now().UnixNano()
	fmt.Printf("rand seed: %v\n", seed)
	rand.Seed(seed)

	var (
		params testutils.E2eParams
	)

	// Deploy the services that you will debug
	BeforeEach(func() {
		testsStarted++
		// Use unique namespaces so we can start tests before namespace is deleted
		// Use predictable namespaces so that we can establish watches
		// (solo-kit does not have a "watch all namespaces" feature yet)
		if os.Getenv("SERIALIZE_NAMESPACES") != "1" {
			testNamespace = fmt.Sprintf("%v-%v", testNSRoot, rand.Int31n(100000))
		} else {
			testNamespace = fmt.Sprintf("%v-%v", testNSRoot, testsStarted)
		}
		params = testutils.NewE2eParams(testNamespace, daName, GinkgoWriter)
		params.SetupE2e()
	})

	// Delete the resources you created
	AfterEach(params.CleanupE2e)

	Describe("Single Container mode", func() {
		It("should get a debug server endpoint", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			time.Sleep(3 * time.Second)
			dbgattachment, err := params.UserController.Attach(daName, params.Namespace, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(9 * time.Second)

			updatedattachment, err := squashcli.WaitCmd(testNamespace, dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())

			// TODO(mitchdraft) put selector spec in a shared package
			nsPods, err := params.KubeClient.CoreV1().Pods(params.Namespace).List(metav1.ListOptions{LabelSelector: sqOpts.SquashLabelSelectorString})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(nsPods.Items)).To(Equal(1))

		})

		It("should get a debug server endpoint, specific process", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.UserController.Attach(daName, params.Namespace, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "service1", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(5 * time.Second)

			updatedattachment, err := squashcli.WaitCmd(testNamespace, dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})

		It("should get a debug server endpoint, specific process that doesn't exist", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.UserController.Attach(daName, params.Namespace, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "processNameDoesntExist", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			updatedattachment, err := squashcli.WaitCmd(testNamespace, dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).NotTo(Equal(core.Status_Accepted))
		})
		It("should attach to two micro services", func() {
			container1 := params.CurrentMicroservicePod.Spec.Containers[0]
			dbgattachment1, err := params.UserController.Attach(daName,
				params.Namespace,
				container1.Image,
				params.CurrentMicroservicePod.ObjectMeta.Name,
				container1.Name,
				"",
				"dlv")
			time.Sleep(5 * time.Second)
			Expect(err).NotTo(HaveOccurred())
			updatedattachment1, err := squashcli.WaitCmd(testNamespace, dbgattachment1.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment1.State).To(Equal(v1.DebugAttachment_Attached))

			container2 := params.Current2MicroservicePod.Spec.Containers[0]
			dbgattachment2, err := params.UserController.Attach(daName2,
				params.Namespace,
				container2.Image,
				params.Current2MicroservicePod.ObjectMeta.Name,
				container2.Name,
				"",
				"dlv")
			time.Sleep(5 * time.Second)
			Expect(err).NotTo(HaveOccurred())
			updatedattachment2, err := squashcli.WaitCmd(testNamespace, dbgattachment2.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment2.State).To(Equal(v1.DebugAttachment_Attached))
		})

		It("should attach and detatch", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.UserController.Attach(daName, params.Namespace, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			testutils.ExpectCounts(params, daName).
				SumPreAttachments(1).
				Attachments(0).
				SumPostAttachments(0)

			time.Sleep(5 * time.Second)

			updatedattachment, err := squashcli.WaitCmd(testNamespace, dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.State).To(Equal(v1.DebugAttachment_Attached))

			testutils.ExpectCounts(params, daName).
				SumPreAttachments(0).
				Attachments(1).
				SumPostAttachments(0)

			dbgattachment, err = params.UserController.RequestDelete(params.Namespace, daName)
			Expect(err).NotTo(HaveOccurred())

			testutils.ExpectCounts(params, daName).
				SumPreAttachments(0).
				Attachments(0).
				SumPostAttachments(1)

			time.Sleep(5 * time.Second)
			testutils.ExpectCounts(params, daName).
				Total(0)
		})

		It("Be able to re-attach once session exited", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.UserController.Attach(daName, params.Namespace, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			time.Sleep(5 * time.Second)
			Expect(err).NotTo(HaveOccurred())
			updatedattachment, err := squashcli.WaitCmd(testNamespace, dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())

			// Ok; now delete the attachment
			dbgattachment, err = params.UserController.RequestDelete(dbgattachment.Metadata.Namespace, dbgattachment.Metadata.Name)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)

			// try again!
			dbgattachment, err = params.UserController.Attach(daName, params.Namespace, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)
			updatedattachment, err = squashcli.WaitCmd(testNamespace, dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})
	})

})
