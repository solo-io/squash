package crd

import (
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/solo-io/squash/pkg/models"
	v1 "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/apis/squash/v1"
	clientset "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/clientset/versioned"
	informers "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/informers/externalversions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/sample-controller/pkg/signals"
)

const (
	namespace = "squash"
)

type CrdClient struct {
	cfg           *restclient.Config
	client        *clientset.Clientset
	controller    *Controller
	apiextsClient *apiexts.Clientset
}

func NewCrdClient(cfg *restclient.Config,
	callback CallbackInterface) (*CrdClient, error) {

	if cfg == nil {
		inconfig, err := rest.InClusterConfig()
		if err != nil {
			panic("not running in a kube cluster")
		}
		cfg = inconfig
	}
	client, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.WithField("error", err).Error("Error in crd.NewCrdClient building squash clientset")
		return nil, err
	}

	stopCh := signals.SetupSignalHandler()

	informerFactory := informers.NewSharedInformerFactory(client /*time.Second*30*/, 0)

	apiClient, err := apiexts.NewForConfig(cfg)
	if err != nil {
		log.WithField("error", err).Error("Error in crd.StartClient building CRD clientset")
		return nil, err
	}

	var controller *Controller
	if callback != nil {
		controller = NewController(apiClient, client, informerFactory, callback)

		go informerFactory.Start(stopCh)
		go func() {
			if err = controller.Run(stopCh); err == nil {
				log.WithField("error", err).Error("Error in crd.StartClient running controller")
				return
			}
		}()
	}

	return &CrdClient{
		cfg:           cfg,
		client:        client,
		controller:    controller,
		apiextsClient: apiClient,
	}, nil
}

func (c *CrdClient) CreateCRDs() error {
	crdAtt := &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: v1.CRDAttachmentsFullName},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   v1.CRDGroup,
			Version: v1.CRDVersion,
			Scope:   v1beta1.NamespaceScoped,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: v1.CRDAttachmentsPlural,
				Kind:   reflect.TypeOf(v1.DebugAttachment{}).Name(),
			},
		},
	}
	if _, err := c.apiextsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crdAtt); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create crd: %v", err)
	}
	crdReq := &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: v1.CRDRequestsFullName},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group:   v1.CRDGroup,
			Version: v1.CRDVersion,
			Scope:   v1beta1.NamespaceScoped,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural: v1.CRDRequestsPlural,
				Kind:   reflect.TypeOf(v1.DebugRequest{}).Name(),
			},
		},
	}
	if _, err := c.apiextsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crdReq); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create crd: %v", err)
	}
	return nil
}

func (c *CrdClient) CreateAttachment(r *models.DebugAttachment, failIfExist bool) (*models.DebugAttachment, error) {
	rv, err := c.client.SquashV1().DebugAttachments(namespace).Create(mapAttachmentToCrd(r, true))
	if err != nil && (failIfExist || !apierrors.IsAlreadyExists(err)) {
		log.WithField("error", err).Error("Error in crd.CreateAttachment")
		return nil, err
	}
	return mapCrdToAttachment(rv), nil
}

func (c *CrdClient) CreateRequest(r *models.DebugRequest, failIfExist bool) (*models.DebugRequest, error) {
	rv, err := c.client.SquashV1().DebugRequests(namespace).Create(mapRequestToCrd(r, true))
	if err != nil && (failIfExist || !apierrors.IsAlreadyExists(err)) {
		log.WithField("error", err).Error("Error in crd.CreateRequest")
		return nil, err
	}
	return mapCrdToRequest(rv), nil
}

func (c *CrdClient) DeleteAttachment(name string) error {
	err := c.client.SquashV1().DebugAttachments(namespace).Delete(name, nil)
	if err != nil {
		log.WithField("error", err).Error("Error in crd.DeleteAttachment")
		return err
	}
	return nil
}

func (c *CrdClient) DeleteRequest(name string) error {
	err := c.client.SquashV1().DebugRequests(namespace).Delete(name, nil)
	if err != nil {
		log.WithField("error", err).Error("Error in crd.DeleteRequest")
		return err
	}
	return nil
}

func (c *CrdClient) UpdateAttachment(r *models.DebugAttachment) (*models.DebugAttachment, error) {
	rv, err := c.client.SquashV1().DebugAttachments(namespace).Update(mapAttachmentToCrd(r, false))
	if err != nil {
		log.WithField("error", err).Error("Error in crd.UpdateAttachment")
		return nil, err
	}
	return mapCrdToAttachment(rv), nil
}

func (c *CrdClient) UpdateRequest(r *models.DebugRequest) (*models.DebugRequest, error) {
	rv, err := c.client.SquashV1().DebugRequests(namespace).Update(mapRequestToCrd(r, false))
	if err != nil {
		log.WithField("error", err).Error("Error in crd.UpdateRequest")
		return nil, err
	}
	return mapCrdToRequest(rv), nil
}

func (c *CrdClient) GetAttachment(name string) (*models.DebugAttachment, error) {
	v, err := c.client.SquashV1().DebugAttachments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.WithField("error", err).Error("Error in crd.GetAttachment")
		return nil, err
	}
	return mapCrdToAttachment(v), nil
}

func (c *CrdClient) GetRequest(name string) (*models.DebugRequest, error) {
	v, err := c.client.SquashV1().DebugRequests(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.WithField("error", err).Error("Error in crd.GetRequest")
		return nil, err
	}
	return mapCrdToRequest(v), nil
}

func (c *CrdClient) ListAttachments() ([]*models.DebugAttachment, error) {
	v, err := c.client.SquashV1().DebugAttachments(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.WithField("error", err).Error("Error in crd.ListAttachments")
		return nil, err
	}

	res := make([]*models.DebugAttachment, 0, len(v.Items))
	for _, r := range v.Items {
		res = append(res, mapCrdToAttachment(&r))
	}
	return res, nil
}

func (c *CrdClient) ListRequests() ([]*models.DebugRequest, error) {
	v, err := c.client.SquashV1().DebugRequests(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.WithField("error", err).Error("Error in crd.ListRequests")
		return nil, err
	}

	res := make([]*models.DebugRequest, 0, len(v.Items))
	for _, r := range v.Items {
		res = append(res, mapCrdToRequest(&r))
	}
	return res, nil
}
