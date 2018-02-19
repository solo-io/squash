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

// FakeDebugRequests implements DebugRequestInterface
type FakeDebugRequests struct {
	Fake *FakeSquashV1
	ns   string
}

var debugrequestsResource = schema.GroupVersionResource{Group: "squash.solo.io", Version: "v1", Resource: "debugrequests"}

var debugrequestsKind = schema.GroupVersionKind{Group: "squash.solo.io", Version: "v1", Kind: "DebugRequest"}

// Get takes name of the debugRequest, and returns the corresponding debugRequest object, and an error if there is any.
func (c *FakeDebugRequests) Get(name string, options v1.GetOptions) (result *squash_v1.DebugRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(debugrequestsResource, c.ns, name), &squash_v1.DebugRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugRequest), err
}

// List takes label and field selectors, and returns the list of DebugRequests that match those selectors.
func (c *FakeDebugRequests) List(opts v1.ListOptions) (result *squash_v1.DebugRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(debugrequestsResource, debugrequestsKind, c.ns, opts), &squash_v1.DebugRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &squash_v1.DebugRequestList{}
	for _, item := range obj.(*squash_v1.DebugRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested debugRequests.
func (c *FakeDebugRequests) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(debugrequestsResource, c.ns, opts))

}

// Create takes the representation of a debugRequest and creates it.  Returns the server's representation of the debugRequest, and an error, if there is any.
func (c *FakeDebugRequests) Create(debugRequest *squash_v1.DebugRequest) (result *squash_v1.DebugRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(debugrequestsResource, c.ns, debugRequest), &squash_v1.DebugRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugRequest), err
}

// Update takes the representation of a debugRequest and updates it. Returns the server's representation of the debugRequest, and an error, if there is any.
func (c *FakeDebugRequests) Update(debugRequest *squash_v1.DebugRequest) (result *squash_v1.DebugRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(debugrequestsResource, c.ns, debugRequest), &squash_v1.DebugRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugRequest), err
}

// Delete takes name of the debugRequest and deletes it. Returns an error if one occurs.
func (c *FakeDebugRequests) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(debugrequestsResource, c.ns, name), &squash_v1.DebugRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeDebugRequests) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(debugrequestsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &squash_v1.DebugRequestList{})
	return err
}

// Patch applies the patch and returns the patched debugRequest.
func (c *FakeDebugRequests) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *squash_v1.DebugRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(debugrequestsResource, c.ns, name, data, subresources...), &squash_v1.DebugRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*squash_v1.DebugRequest), err
}
