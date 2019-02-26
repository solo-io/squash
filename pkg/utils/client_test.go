package utils_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/test/testutils"
)

var _ = Describe("utils", func() {
	It("should generate debug attachment", func() {
		ctx := context.TODO()
		daClient, err := utils.GetDebugAttachmentClient(ctx)
		Expect(err).To(BeNil())

		name := "aname2"
		namespace := "default"
		da := testutils.GenerateDebugAttachmentDlv1(name, namespace)
		writeOpts := clients.WriteOpts{
			Ctx:               ctx,
			OverwriteExisting: true,
		}
		written, err := daClient.Write(&da, writeOpts)
		Expect(err).To(BeNil())
		Expect(written.Metadata.Name).To(Equal(name))
		Expect(written.Metadata.Namespace).To(Equal(namespace))

		readOpts := clients.ReadOpts{
			Ctx: ctx,
		}
		read, err := daClient.Read(namespace, name, readOpts)
		Expect(err).To(BeNil())
		Expect(read.Metadata.Name).To(Equal(name))
		Expect(read.Metadata.Namespace).To(Equal(namespace))

		// Cleanup
		deleteOpts := clients.DeleteOpts{
			Ctx:            ctx,
			IgnoreNotExist: false,
		}
		err = daClient.Delete(namespace, name, deleteOpts)
		Expect(err).To(BeNil())
	})
})
