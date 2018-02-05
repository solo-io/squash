// Squash by Solo.io

package fake

import (
	v1 "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/clientset/versioned/typed/squash/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeSquashV1 struct {
	*testing.Fake
}

func (c *FakeSquashV1) DebugAttachments(namespace string) v1.DebugAttachmentInterface {
	return &FakeDebugAttachments{c, namespace}
}

func (c *FakeSquashV1) DebugRequests(namespace string) v1.DebugRequestInterface {
	return &FakeDebugRequests{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeSquashV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
