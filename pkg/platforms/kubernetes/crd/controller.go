package crd

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/apis/squash/v1"
	clientset "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/clientset/versioned"
	informers "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/informers/externalversions"
	listers "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/client/listers/squash/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	TypeDebugAttachment = iota
	TypeDebugRequest

	ActionAdd = iota
	ActionUpdate
	ActionDelete
)

type CallbackInterface interface {
	SyncCRDChanges(objtype ObjectType, action Action, obj interface{})
}

type ObjectType int
type Action int

type Controller struct {
	crdClientset      apiexts.Interface
	squashClientset   clientset.Interface
	attachmentsLister listers.DebugAttachmentLister
	attachmentsSynced cache.InformerSynced
	requestsLister    listers.DebugRequestLister
	requestsSynced    cache.InformerSynced
	workQueue         workqueue.RateLimitingInterface
	callback          CallbackInterface
}

type WorkItem struct {
	ot  ObjectType
	act Action
	obj interface{}
}

func NewController(
	crdClientset apiexts.Interface,
	squashclientset clientset.Interface,
	sqinformer informers.SharedInformerFactory,
	callback CallbackInterface) *Controller {

	dai := sqinformer.Squash().V1().DebugAttachments()
	dri := sqinformer.Squash().V1().DebugRequests()

	controller := &Controller{
		crdClientset:      crdClientset,
		squashClientset:   squashclientset,
		attachmentsLister: dai.Lister(),
		attachmentsSynced: dai.Informer().HasSynced,
		requestsLister:    dri.Lister(),
		requestsSynced:    dri.Informer().HasSynced,
		workQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SquashCRDQueue"),
		callback:          callback,
	}

	dai.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addAttachment,
		DeleteFunc: controller.deleteAttachment,
		UpdateFunc: controller.updateAttachment,
	})

	dri.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.addRequest,
		DeleteFunc: controller.deleteRequest,
		UpdateFunc: controller.updateRequest,
	})
	return controller
}

func NewAttachmentItem(action Action, obj interface{}) *WorkItem {
	w, ok := obj.(*v1.DebugAttachment)
	if ok {
		return &WorkItem{
			ot:  TypeDebugAttachment,
			act: action,
			obj: mapCrdToAttachment(w),
		}
	} else {
		log.WithField("obj", obj).Error("Bad attachment")
	}
	return nil
}

func NewRequestItem(action Action, obj interface{}) *WorkItem {
	w, ok := obj.(*v1.DebugRequest)
	if ok {
		return &WorkItem{
			ot:  TypeDebugRequest,
			act: action,
			obj: mapCrdToRequest(w),
		}
	} else {
		log.WithField("obj", obj).Error("Bad request")
	}
	return nil
}

func (c *Controller) addAttachment(obj interface{}) {
	w := NewAttachmentItem(ActionAdd, obj)
	if w != nil {
		c.workQueue.AddRateLimited(w)
	}
}

func (c *Controller) updateAttachment(oldobj, obj interface{}) {
	w := NewAttachmentItem(ActionUpdate, obj)
	if w != nil {
		c.workQueue.AddRateLimited(w)
	}
}

func (c *Controller) deleteAttachment(obj interface{}) {
	w := NewAttachmentItem(ActionDelete, obj)
	if w != nil {
		c.workQueue.AddRateLimited(w)
	}
}

func (c *Controller) addRequest(obj interface{}) {
	w := NewRequestItem(ActionAdd, obj)
	if w != nil {
		c.workQueue.AddRateLimited(w)
	}
}

func (c *Controller) updateRequest(oldobj, obj interface{}) {
	w := NewRequestItem(ActionUpdate, obj)
	if w != nil {
		c.workQueue.AddRateLimited(w)
	}
}

func (c *Controller) deleteRequest(obj interface{}) {
	w := NewRequestItem(ActionDelete, obj)
	if w != nil {
		c.workQueue.AddRateLimited(w)
	}
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workQueue.ShutDown()

	if ok := cache.WaitForCacheSync(stopCh, c.attachmentsSynced, c.requestsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	go wait.Until(c.runWorkerOnQueue, time.Second, stopCh)

	<-stopCh

	return nil
}

func (c *Controller) runWorkerOnQueue() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workQueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workQueue.Done(obj)
		wi, ok := obj.(*WorkItem)
		if ok && c.callback != nil {
			c.callback.SyncCRDChanges(wi.ot, wi.act, wi.obj)
		}
		c.workQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}
	return true
}
