// +k8s:deepcopy-gen=package

// +groupName=squash.solo.io
package v1

const (
	CRDAttachmentsPlural   = "debugattachments"
	CRDRequestsPlural      = "debugrequests"
	CRDGroup               = "squash.solo.io"
	CRDVersion             = "v1"
	CRDAttachmentsFullName = CRDAttachmentsPlural + "." + CRDGroup
	CRDRequestsFullName    = CRDRequestsPlural + "." + CRDGroup
)
