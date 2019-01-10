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
func (c *Counter) RequestingAttachments(val int) *Counter {
	Expect(c.counts[v1.DebugAttachment_RequestingAttachment]).To(Equal(val))
	return c
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
func (c *Counter) RequestingDeletes(val int) *Counter {
	Expect(c.counts[v1.DebugAttachment_RequestingDelete]).To(Equal(val))
	return c
}

func (c *Counter) SumPreAttachments(val int) *Counter {
	preAttCount := c.counts[v1.DebugAttachment_RequestingAttachment] +
		c.counts[v1.DebugAttachment_PendingAttachment]
	Expect(preAttCount).To(Equal(val))
	return c
}
func (c *Counter) SumPostAttachments(val int) *Counter {
	postAttCount := c.counts[v1.DebugAttachment_PendingDelete] +
		c.counts[v1.DebugAttachment_RequestingDelete]
	Expect(postAttCount).To(Equal(val))
	return c
}
