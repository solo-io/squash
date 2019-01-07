package e2e_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	squashcli "github.com/solo-io/squash/pkg/cmd/cli"
	"github.com/solo-io/squash/test/testutils"
)

//const KubeEndpoint = "http://localhost:8001/api"

func Must(err error) {
	Expect(err).NotTo(HaveOccurred())
}

var daName = "mitch"
var daName2 = "mitch2"

var _ = Describe("Single debug mode", func() {

	var (
		params testutils.E2eParams
	)

	/*
		Deploy the services that you will debug

	*/
	BeforeEach(func() {
		params = testutils.NewE2eParams(daName)
		params.SetupE2e()
	})

	AfterEach(params.Cleanup)

	Describe("Single Container mode", func() {
		It("should get a debug server endpoint", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.Squash.Attach(daName, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})

		It("should get a debug server endpoint, specific process", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.Squash.Attach(daName, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "service1", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)

			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})

		It("should get a debug server endpoint, specific process that doesn't exist", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.Squash.Attach(daName, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "processNameDoesntExist", "dlv")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(time.Second)

			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).NotTo(Equal(core.Status_Accepted))
		})
		FIt("should attach to two micro services", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.Squash.Attach(daName, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).To(Equal(core.Status_Accepted))

			container = params.Current2MicroservicePod.Spec.Containers[0]
			dbgattachment, err = params.Squash.Attach(daName2, container.Image, params.Current2MicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(time.Second)
			updatedattachment, err = squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.Status.State).To(Equal(core.Status_Accepted))
		})

		It("Be able to re-attach once session exited", func() {
			container := params.CurrentMicroservicePod.Spec.Containers[0]

			dbgattachment, err := params.Squash.Attach(daName, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			updatedattachment, err := squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())

			// Ok; now delete the attachment
			err = params.Squash.Delete(dbgattachment)
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(5 * time.Second)

			// try again!
			dbgattachment, err = params.Squash.Attach(daName, container.Image, params.CurrentMicroservicePod.ObjectMeta.Name, container.Name, "", "dlv")
			Expect(err).NotTo(HaveOccurred())
			updatedattachment, err = squashcli.WaitCmd(dbgattachment.Metadata.Name, 1.0)

			Expect(err).NotTo(HaveOccurred())
			Expect(updatedattachment.DebugServerAddress).ToNot(BeEmpty())
		})
	})

})
