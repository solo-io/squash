package testutils

import (
	. "github.com/onsi/gomega"

	"github.com/solo-io/squash/pkg/api/v1"
)

type Counter struct {
	counts map[v1.DebugAttachment_State]int
}

func ExpectCounts(params E2eParams, name string) *Counter {
	counts, err := params.UserController.Counts(params.Namespace, name)
	Expect(err).NotTo(HaveOccurred())
	return &Counter{counts: counts}
}
func (c *Counter) PendingAttachments(val int) *Counter {
	Expect(c.counts[v1.DebugAttachment_PendingAttachment]).To(Equal(val))
	return c
}
func (c *Counter) Attachments(val int) *Counter {
	Expect(c.counts[v1.DebugAttachment_Attached]).To(Equal(val))
	return c
}
func (c *Counter) PendingDeletes(val int) *Counter {
	Expect(c.counts[v1.DebugAttachment_PendingDelete]).To(Equal(val))
	return c
}
