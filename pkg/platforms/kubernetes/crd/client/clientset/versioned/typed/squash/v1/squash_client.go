// Squash by Solo.io

package v1

import (
	v1 "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/apis/squash/v1"
	"github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/clientset/versioned/scheme"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	rest "k8s.io/client-go/rest"
)

type SquashV1Interface interface {
	RESTClient() rest.Interface
	DebugAttachmentsGetter
	DebugRequestsGetter
}

// SquashV1Client is used to interact with features provided by the squash.solo.io group.
type SquashV1Client struct {
	restClient rest.Interface
}

func (c *SquashV1Client) DebugAttachments(namespace string) DebugAttachmentInterface {
	return newDebugAttachments(c, namespace)
}

func (c *SquashV1Client) DebugRequests(namespace string) DebugRequestInterface {
	return newDebugRequests(c, namespace)
}

// NewForConfig creates a new SquashV1Client for the given config.
func NewForConfig(c *rest.Config) (*SquashV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &SquashV1Client{client}, nil
}

// NewForConfigOrDie creates a new SquashV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *SquashV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new SquashV1Client for the given RESTClient.
func New(c rest.Interface) *SquashV1Client {
	return &SquashV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *SquashV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
