package util_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/test/util"
)

func generateDebugAttachment(name, namespace, dbgger, image, pod, container, processName string) v1.DebugAttachment {

	da := v1.DebugAttachment{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
		Debugger:  dbgger,
		Image:     image,
		Pod:       pod,
		Container: container,
	}
	if processName != "" {
		da.ProcessName = processName
	}
	return da
}
func generateDebugAttachmentDlv1(name, namespace string) v1.DebugAttachment {
	dbgger := "dlv"
	image := "mk"
	pod := "somepod"
	container := "somecontainer"
	processName := "pcsnm"
	return generateDebugAttachment(name, namespace, dbgger, image, pod, container, processName)
}

var _ = Describe("utils", func() {
	It("should generate debug attachment", func() {
		ctx := context.TODO()
		daClient, err := util.GetDebugAttachmentClient(ctx)
		Expect(err).To(BeNil())

		name := "aname2"
		namespace := options.SquashNamespace
		da := generateDebugAttachmentDlv1(name, namespace)
		writeOpts := clients.WriteOpts{
			Ctx:               ctx,
			OverwriteExisting: true,
		}
		written, err := (*daClient).Write(&da, writeOpts)
		Expect(err).To(BeNil())
		Expect(written.Metadata.Name).To(Equal(name))
		Expect(written.Metadata.Namespace).To(Equal(namespace))

		readOpts := clients.ReadOpts{
			Ctx: ctx,
		}
		read, err := (*daClient).Read(namespace, name, readOpts)
		Expect(err).To(BeNil())
		Expect(read.Metadata.Name).To(Equal(name))
		Expect(read.Metadata.Namespace).To(Equal(namespace))
		fmt.Println(read)

		// Cleanup
		deleteOpts := clients.DeleteOpts{
			Ctx:            ctx,
			IgnoreNotExist: false,
		}
		err = (*daClient).Delete(namespace, name, deleteOpts)
		Expect(err).To(BeNil())
	})
})
