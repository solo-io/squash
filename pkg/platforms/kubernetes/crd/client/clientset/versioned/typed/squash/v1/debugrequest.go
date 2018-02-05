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

// DebugRequestsGetter has a method to return a DebugRequestInterface.
// A group's client should implement this interface.
type DebugRequestsGetter interface {
	DebugRequests(namespace string) DebugRequestInterface
}

// DebugRequestInterface has methods to work with DebugRequest resources.
type DebugRequestInterface interface {
	Create(*v1.DebugRequest) (*v1.DebugRequest, error)
	Update(*v1.DebugRequest) (*v1.DebugRequest, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.DebugRequest, error)
	List(opts meta_v1.ListOptions) (*v1.DebugRequestList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.DebugRequest, err error)
	DebugRequestExpansion
}

// debugRequests implements DebugRequestInterface
type debugRequests struct {
	client rest.Interface
	ns     string
}

// newDebugRequests returns a DebugRequests
func newDebugRequests(c *SquashV1Client, namespace string) *debugRequests {
	return &debugRequests{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the debugRequest, and returns the corresponding debugRequest object, and an error if there is any.
func (c *debugRequests) Get(name string, options meta_v1.GetOptions) (result *v1.DebugRequest, err error) {
	result = &v1.DebugRequest{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("debugrequests").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of DebugRequests that match those selectors.
func (c *debugRequests) List(opts meta_v1.ListOptions) (result *v1.DebugRequestList, err error) {
	result = &v1.DebugRequestList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("debugrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested debugRequests.
func (c *debugRequests) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("debugrequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a debugRequest and creates it.  Returns the server's representation of the debugRequest, and an error, if there is any.
func (c *debugRequests) Create(debugRequest *v1.DebugRequest) (result *v1.DebugRequest, err error) {
	result = &v1.DebugRequest{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("debugrequests").
		Body(debugRequest).
		Do().
		Into(result)
	return
}

// Update takes the representation of a debugRequest and updates it. Returns the server's representation of the debugRequest, and an error, if there is any.
func (c *debugRequests) Update(debugRequest *v1.DebugRequest) (result *v1.DebugRequest, err error) {
	result = &v1.DebugRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("debugrequests").
		Name(debugRequest.Name).
		Body(debugRequest).
		Do().
		Into(result)
	return
}

// Delete takes name of the debugRequest and deletes it. Returns an error if one occurs.
func (c *debugRequests) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("debugrequests").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *debugRequests) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("debugrequests").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched debugRequest.
func (c *debugRequests) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.DebugRequest, err error) {
	result = &v1.DebugRequest{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("debugrequests").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
