// Squash by Solo.io

package fake

import (
	squash_v1 "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/apis/squash/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeDebugAttachments implements DebugAttachmentInterface
type FakeDebugAttachments struct {
	Fake *FakeSquashV1
	ns   string
}

var debugattachmentsResource = schema.GroupVersionResource{Group: "squash.solo.io", Version: "v1", Resource: "debugattachments"}

var debugattachmentsKind = schema.GroupVersionKind{Group: "squash.solo.io", Version: "v1", Kind: "DebugAttachment"}

// Get takes name of the debugAttachment, and returns the corresponding debugAttachment object, and an error if there is any.
func (c *FakeDebugAttachments) Get(name string, options v1.GetOptions) (result *squash_v1.DebugAttachment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(debugattachmentsResource, c.ns, name), &squash_v1.DebugAttachment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugAttachment), err
}

// List takes label and field selectors, and returns the list of DebugAttachments that match those selectors.
func (c *FakeDebugAttachments) List(opts v1.ListOptions) (result *squash_v1.DebugAttachmentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(debugattachmentsResource, debugattachmentsKind, c.ns, opts), &squash_v1.DebugAttachmentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &squash_v1.DebugAttachmentList{}
	for _, item := range obj.(*squash_v1.DebugAttachmentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested debugAttachments.
func (c *FakeDebugAttachments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(debugattachmentsResource, c.ns, opts))

}

// Create takes the representation of a debugAttachment and creates it.  Returns the server's representation of the debugAttachment, and an error, if there is any.
func (c *FakeDebugAttachments) Create(debugAttachment *squash_v1.DebugAttachment) (result *squash_v1.DebugAttachment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(debugattachmentsResource, c.ns, debugAttachment), &squash_v1.DebugAttachment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugAttachment), err
}

// Update takes the representation of a debugAttachment and updates it. Returns the server's representation of the debugAttachment, and an error, if there is any.
func (c *FakeDebugAttachments) Update(debugAttachment *squash_v1.DebugAttachment) (result *squash_v1.DebugAttachment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(debugattachmentsResource, c.ns, debugAttachment), &squash_v1.DebugAttachment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugAttachment), err
}

// Delete takes name of the debugAttachment and deletes it. Returns an error if one occurs.
func (c *FakeDebugAttachments) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(debugattachmentsResource, c.ns, name), &squash_v1.DebugAttachment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDebugAttachments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(debugattachmentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &squash_v1.DebugAttachmentList{})
	return err
}

// Patch applies the patch and returns the patched debugAttachment.
func (c *FakeDebugAttachments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *squash_v1.DebugAttachment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(debugattachmentsResource, c.ns, name, data, subresources...), &squash_v1.DebugAttachment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugAttachment), err
}
