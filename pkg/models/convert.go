package models

import "github.com/solo-io/squash/pkg/api/v1"

func ConvertDebugAttachments(das []*v1.DebugAttachment) []DebugAttachment {
	out := []DebugAttachment{}
	for _, da := range das {
		out = append(out, ConvertDebugAttachment(da))
	}
	return out
}

func ConvertDebugAttachment(da *v1.DebugAttachment) DebugAttachment {
	// for compatibility with the current version of vs code squash v0.1.9
	state := da.State.String()
	if state == v1.DebugAttachment_Attached.String() {
		state = "attached"
	}
	return DebugAttachment{
		Metadata: &ObjectMeta{
			Name: da.Metadata.Name,
		},
		Spec: &DebugAttachmentSpec{
			Attachment:   da.Attachment,
			Debugger:     da.Debugger,
			Image:        da.Image,
			MatchRequest: da.MatchRequest,
			Node:         da.Node,
			ProcessName:  da.ProcessName,
		},
		Status: &DebugAttachmentStatus{
			DebugServerAddress: da.DebugServerAddress,
			State:              state,
		},
	}
}

func ConvertDebugRequests(drs []*v1.DebugAttachment) []DebugRequest {
	out := []DebugRequest{}
	for _, dr := range drs {
		out = append(out, ConvertDebugRequest(dr))
	}
	return out
}

func ConvertDebugRequest(dr *v1.DebugAttachment) DebugRequest {
	return DebugRequest{
		Metadata: &ObjectMeta{
			Name: dr.Metadata.Name,
		},
		Spec: &DebugRequestSpec{
			Debugger:    &dr.Debugger,
			Image:       &dr.Image,
			ProcessName: dr.ProcessName,
		},
		Status: &DebugRequestStatus{
			DebugAttachmentRef: dr.DebugServerAddress,
		},
	}
}
