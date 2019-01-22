package cli

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
)

func (o *Options) getAllDebugAttachments() (v1.DebugAttachmentList, error) {
	kubeResClient, err := kubeutils.NewOutOfClusterKubeClientset()
	if err != nil {
		return v1.DebugAttachmentList{}, err
	}
	watchNamespaces, err := kubeutils.GetNamespaces(kubeResClient)
	if err != nil {
		return v1.DebugAttachmentList{}, err
	}
	das := v1.DebugAttachmentList{}
	for _, ns := range watchNamespaces {
		nsDas, err := (*o.daClient).List(ns, clients.ListOpts{Ctx: o.ctx})
		if err != nil {
			return v1.DebugAttachmentList{}, err
		}
		for _, nsDa := range nsDas {
			das = append(das, nsDa)
		}
	}
	return das, nil
}

func (o *Options) getNamedDebugAttachment(name string) (*v1.DebugAttachment, error) {
	das, err := o.getAllDebugAttachments()
	if err != nil {
		return &v1.DebugAttachment{}, err
	}

	namedDas := v1.DebugAttachmentList{}
	for _, nDa := range das {
		if nDa.Metadata.Name == name {
			namedDas = append(namedDas, nDa)
		}
	}
	if len(namedDas) > 1 {
		// TODO(mitchdraft) - make this impossible by explicitly specifying the namespace
		return &v1.DebugAttachment{}, fmt.Errorf("multiple debug attachments with the same name found")
	}
	if len(namedDas) == 0 {
		return &v1.DebugAttachment{}, fmt.Errorf("Debug attachment %v not found", name)
	}
	return namedDas[0], nil
}
