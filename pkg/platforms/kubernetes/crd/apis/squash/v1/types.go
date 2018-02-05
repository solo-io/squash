package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DebugAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec
	// Required: true
	Spec *DebugAttachmentSpec `json:"spec"`

	// status
	Status *DebugAttachmentStatus `json:"status,omitempty"`
}

type DebugAttachmentSpec struct {

	// attachment
	// Required: true
	Attachment *KubeAttachment `json:"attachment"`

	// debugger
	Debugger string `json:"debugger,omitempty"`

	// image
	Image string `json:"image,omitempty"`

	// If true, this attachment must match a pending debug request.
	MatchRequest bool `json:"match_request,omitempty"`

	// node
	Node string `json:"node,omitempty"`

	// process name
	ProcessName string `json:"process_name,omitempty"`
}

type DebugAttachmentStatus struct {

	// debug server address
	DebugServerAddress string `json:"debug_server_address,omitempty"`

	// state
	State string `json:"state,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DebugAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DebugAttachment `json:"items"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DebugRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec
	// Required: true
	Spec *DebugRequestSpec `json:"spec"`

	// status
	Status *DebugRequestStatus `json:"status,omitempty"`
}

type DebugRequestSpec struct {

	// debugger
	// Required: true
	Debugger *string `json:"debugger"`

	// image
	// Required: true
	Image *string `json:"image"`

	// process name
	ProcessName string `json:"process_name,omitempty"`
}

type DebugRequestStatus struct {

	// debug attachment ref
	DebugAttachmentRef string `json:"debug_attachment_ref,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DebugRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DebugRequest `json:"items"`
}

type KubeAttachment struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}
