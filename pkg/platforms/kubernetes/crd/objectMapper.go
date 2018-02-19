package crd

import (
	"github.com/solo-io/squash/pkg/models"
	v1 "github.com/solo-io/squash/pkg/platforms/kubernetes/crd/apis/squash/v1"
	k8models "github.com/solo-io/squash/pkg/platforms/kubernetes/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func mapAttachmentToCrd(r *models.DebugAttachment, isSkipMeta bool) *v1.DebugAttachment {
	meta := metav1.ObjectMeta{GenerateName: "a"}
	if !isSkipMeta {
		meta = metav1.ObjectMeta{
			Name:            r.Metadata.Name,
			ResourceVersion: r.Metadata.Version,
		}
	}
	var spec = &v1.DebugAttachmentSpec{}
	if r.Spec != nil {
		var ka *v1.KubeAttachment
		if r.Spec.Attachment != nil {
			k := r.Spec.Attachment.(*k8models.KubeAttachment)
			ka = &v1.KubeAttachment{
				Namespace: k.Namespace,
				Pod:       k.Pod,
				Container: k.Container,
			}
		}

		spec = &v1.DebugAttachmentSpec{
			Attachment:   ka,
			Debugger:     r.Spec.Debugger,
			Image:        r.Spec.Image,
			MatchRequest: r.Spec.MatchRequest,
			Node:         r.Spec.Node,
			ProcessName:  r.Spec.ProcessName,
		}
	} else {
		spec = nil
	}

	var status = &v1.DebugAttachmentStatus{}

	if r.Status != nil {
		status = &v1.DebugAttachmentStatus{
			DebugServerAddress: r.Status.DebugServerAddress,
			State:              r.Status.State,
		}
	} else {
		status = nil
	}

	return &v1.DebugAttachment{
		ObjectMeta: meta,
		Spec:       spec,
		Status:     status,
	}
}

func mapRequestToCrd(r *models.DebugRequest, isSkipMeta bool) *v1.DebugRequest {
	meta := metav1.ObjectMeta{GenerateName: "r"}
	if !isSkipMeta {
		meta = metav1.ObjectMeta{
			Name:            r.Metadata.Name,
			ResourceVersion: r.Metadata.Version,
		}
	}
	var spec = &v1.DebugRequestSpec{}

	if r.Spec != nil {
		spec = &v1.DebugRequestSpec{
			Debugger:    r.Spec.Debugger,
			Image:       r.Spec.Image,
			ProcessName: r.Spec.ProcessName,
		}
	} else {
		spec = nil
	}

	var status = &v1.DebugRequestStatus{}
	if r.Status != nil {
		status = &v1.DebugRequestStatus{
			DebugAttachmentRef: r.Status.DebugAttachmentRef,
		}
	} else {
		status = nil
	}

	return &v1.DebugRequest{
		ObjectMeta: meta,
		Spec:       spec,
		Status:     status,
	}
}

func mapCrdToAttachment(r *v1.DebugAttachment) *models.DebugAttachment {
	meta := &models.ObjectMeta{
		Name:    r.ObjectMeta.Name,
		Version: r.ObjectMeta.ResourceVersion,
	}

	var spec = &models.DebugAttachmentSpec{}
	if r.Spec != nil {
		var ka *k8models.KubeAttachment
		if r.Spec.Attachment != nil {
			ka = &k8models.KubeAttachment{
				Namespace: r.Spec.Attachment.Namespace,
				Pod:       r.Spec.Attachment.Pod,
				Container: r.Spec.Attachment.Container,
			}
		}

		spec = &models.DebugAttachmentSpec{
			Attachment:   ka,
			Debugger:     r.Spec.Debugger,
			Image:        r.Spec.Image,
			MatchRequest: r.Spec.MatchRequest,
			Node:         r.Spec.Node,
			ProcessName:  r.Spec.ProcessName,
		}
	} else {
		spec = nil
	}

	var status = &models.DebugAttachmentStatus{}

	if r.Status != nil {
		status = &models.DebugAttachmentStatus{
			DebugServerAddress: r.Status.DebugServerAddress,
			State:              r.Status.State,
		}
	} else {
		status = nil
	}

	return &models.DebugAttachment{
		Metadata: meta,
		Spec:     spec,
		Status:   status,
	}
}

func mapCrdToRequest(r *v1.DebugRequest) *models.DebugRequest {
	meta := &models.ObjectMeta{
		Name:    r.ObjectMeta.Name,
		Version: r.ObjectMeta.ResourceVersion,
	}
	var spec = &models.DebugRequestSpec{}

	if r.Spec != nil {
		spec = &models.DebugRequestSpec{
			Debugger:    r.Spec.Debugger,
			Image:       r.Spec.Image,
			ProcessName: r.Spec.ProcessName,
		}
	} else {
		spec = nil
	}

	var status = &models.DebugRequestStatus{}
	if r.Status != nil {
		status = &models.DebugRequestStatus{
			DebugAttachmentRef: r.Status.DebugAttachmentRef,
		}
	} else {
		status = nil
	}

	return &models.DebugRequest{
		Metadata: meta,
		Spec:     spec,
		Status:   status,
	}
}
