// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type TransitionDebugAttachmentFunc func(original, desired *DebugAttachment) (bool, error)

type DebugAttachmentReconciler interface {
	Reconcile(namespace string, desiredResources DebugAttachmentList, transition TransitionDebugAttachmentFunc, opts clients.ListOpts) error
}

func debugAttachmentsToResources(list DebugAttachmentList) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, debugAttachment := range list {
		resourceList = append(resourceList, debugAttachment)
	}
	return resourceList
}

func NewDebugAttachmentReconciler(client DebugAttachmentClient) DebugAttachmentReconciler {
	return &debugAttachmentReconciler{
		base: reconcile.NewReconciler(client.BaseClient()),
	}
}

type debugAttachmentReconciler struct {
	base reconcile.Reconciler
}

func (r *debugAttachmentReconciler) Reconcile(namespace string, desiredResources DebugAttachmentList, transition TransitionDebugAttachmentFunc, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "debugAttachment_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*DebugAttachment), desired.(*DebugAttachment))
		}
	}
	return r.base.Reconcile(namespace, debugAttachmentsToResources(desiredResources), transitionResources, opts)
}
