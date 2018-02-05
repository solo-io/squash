// Squash by Solo.io

package v1

import (
	v1 "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/apis/squash/v1"
	scheme "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// DebugAttachmentsGetter has a method to return a DebugAttachmentInterface.
// A group's client should implement this interface.
type DebugAttachmentsGetter interface {
	DebugAttachments(namespace string) DebugAttachmentInterface
}

// DebugAttachmentInterface has methods to work with DebugAttachment resources.
type DebugAttachmentInterface interface {
	Create(*v1.DebugAttachment) (*v1.DebugAttachment, error)
	Update(*v1.DebugAttachment) (*v1.DebugAttachment, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.DebugAttachment, error)
	List(opts meta_v1.ListOptions) (*v1.DebugAttachmentList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.DebugAttachment, err error)
	DebugAttachmentExpansion
}

// debugAttachments implements DebugAttachmentInterface
type debugAttachments struct {
	client rest.Interface
	ns     string
}

// newDebugAttachments returns a DebugAttachments
func newDebugAttachments(c *SquashV1Client, namespace string) *debugAttachments {
	return &debugAttachments{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the debugAttachment, and returns the corresponding debugAttachment object, and an error if there is any.
func (c *debugAttachments) Get(name string, options meta_v1.GetOptions) (result *v1.DebugAttachment, err error) {
	result = &v1.DebugAttachment{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("debugattachments").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DebugAttachments that match those selectors.
func (c *debugAttachments) List(opts meta_v1.ListOptions) (result *v1.DebugAttachmentList, err error) {
	result = &v1.DebugAttachmentList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("debugattachments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested debugAttachments.
func (c *debugAttachments) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("debugattachments").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a debugAttachment and creates it.  Returns the server's representation of the debugAttachment, and an error, if there is any.
func (c *debugAttachments) Create(debugAttachment *v1.DebugAttachment) (result *v1.DebugAttachment, err error) {
	result = &v1.DebugAttachment{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("debugattachments").
		Body(debugAttachment).
		Do().
		Into(result)
	return
}

// Update takes the representation of a debugAttachment and updates it. Returns the server's representation of the debugAttachment, and an error, if there is any.
func (c *debugAttachments) Update(debugAttachment *v1.DebugAttachment) (result *v1.DebugAttachment, err error) {
	result = &v1.DebugAttachment{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("debugattachments").
		Name(debugAttachment.Name).
		Body(debugAttachment).
		Do().
		Into(result)
	return
}

// Delete takes name of the debugAttachment and deletes it. Returns an error if one occurs.
func (c *debugAttachments) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("debugattachments").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *debugAttachments) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("debugattachments").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched debugAttachment.
func (c *debugAttachments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.DebugAttachment, err error) {
	result = &v1.DebugAttachment{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("debugattachments").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
