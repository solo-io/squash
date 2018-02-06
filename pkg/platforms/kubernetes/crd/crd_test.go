package crd_test

import (
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/platforms/kubernetes/crd"
	k8models "github.com/solo-io/squash/pkg/platforms/kubernetes/models"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kubecfgsfx = ".kube/config"
	master     = ""
	dbgName    = "testDebugger"
	badDbgName = "badDebugger"
	numInList  = 10
)

var crdClient *crd.CrdClient

// CRD callback
func cb(objtype crd.ObjectType, action crd.Action, obj interface{}) {
	if objtype == crd.TypeDebugAttachment {
		v, ok := obj.(*models.DebugAttachment)
		Expect(ok).To(BeTrue())
		Expect(v).ToNot(BeNil())
		Expect(v.Metadata).ToNot(BeNil())
	} else {
		v, ok := obj.(*models.DebugRequest)
		Expect(ok).To(BeTrue())
		Expect(v).ToNot(BeNil())
		Expect(v.Metadata).ToNot(BeNil())
	}
}

func init() {
	RegisterFailHandler(Fail)
	var err error
	crdClient, err = initTests(cb)
	Expect(err).NotTo(HaveOccurred())
}

func TestAttachment(t *testing.T) {
	Expect(crdClient).NotTo(BeNil())

	a, err := createAttachment()
	Expect(err).NotTo(HaveOccurred())

	a.Spec.Debugger = dbgName
	a, err = crdClient.UpdateAttachment(a)
	Expect(err).NotTo(HaveOccurred())

	ab, err := crdClient.GetAttachment(a.Metadata.Name)
	Expect(err).NotTo(HaveOccurred())

	Expect(a.Spec.Debugger == ab.Spec.Debugger).To(BeTrue())

	_, ok := a.Spec.Attachment.(*k8models.KubeAttachment)
	Expect(ok).To(BeTrue())

	err = crdClient.DeleteAttachment(a.Metadata.Name)
	Expect(err).NotTo(HaveOccurred())
}

func TestRequest(t *testing.T) {
	Expect(crdClient).NotTo(BeNil())

	r, err := createRequest()
	Expect(err).NotTo(HaveOccurred())

	dbgNameGood := dbgName
	r.Spec.Debugger = &dbgNameGood

	r, err = crdClient.UpdateRequest(r)
	Expect(err).NotTo(HaveOccurred())

	rb, err := crdClient.GetRequest(r.Metadata.Name)
	Expect(err).NotTo(HaveOccurred())

	Expect(*r.Spec.Debugger == *rb.Spec.Debugger).To(BeTrue())

	err = crdClient.DeleteRequest(r.Metadata.Name)
	Expect(err).NotTo(HaveOccurred())
}

func TestAttachmentsList(t *testing.T) {
	Expect(crdClient).NotTo(BeNil())

	names := make(map[string]bool)
	for i := 0; i < numInList; i++ {
		r, err := createAttachment()
		Expect(err).NotTo(HaveOccurred())
		names[r.Metadata.Name] = true
	}

	v, err := crdClient.ListAttachments()
	Expect(err).NotTo(HaveOccurred())

	for _, rv := range v {
		if _, ok := names[rv.Metadata.Name]; ok {
			err = crdClient.DeleteAttachment(rv.Metadata.Name)
			Expect(err).NotTo(HaveOccurred())
			delete(names, rv.Metadata.Name)
		}
	}
	Expect(len(names)).To(BeZero())
}

func TestRequestsList(t *testing.T) {
	Expect(crdClient).NotTo(BeNil())

	names := make(map[string]bool)
	for i := 0; i < numInList; i++ {
		r, err := createRequest()
		Expect(err).NotTo(HaveOccurred())
		names[r.Metadata.Name] = true
	}

	v, err := crdClient.ListRequests()
	Expect(err).NotTo(HaveOccurred())

	for _, rv := range v {
		if _, ok := names[rv.Metadata.Name]; ok {
			err = crdClient.DeleteRequest(rv.Metadata.Name)
			Expect(err).NotTo(HaveOccurred())
			delete(names, rv.Metadata.Name)
		}
	}
	Expect(len(names)).To(BeZero())
}

func createAttachment() (*models.DebugAttachment, error) {

	ka := &k8models.KubeAttachment{
		Pod: "testPod",
	}

	a := &models.DebugAttachment{
		Metadata: &models.ObjectMeta{Name: ""},
		Spec: &models.DebugAttachmentSpec{
			Attachment: ka,
			Debugger:   badDbgName,
		},
	}
	return crdClient.CreateAttachment(a, false)
}

func createRequest() (*models.DebugRequest, error) {
	dbgNameBad := badDbgName
	r := &models.DebugRequest{
		Metadata: &models.ObjectMeta{Name: ""},
		Spec: &models.DebugRequestSpec{
			Debugger: &dbgNameBad,
		},
	}
	return crdClient.CreateRequest(r, false)
}

func initTests(cb func(objtype crd.ObjectType, action crd.Action, obj interface{})) (*crd.CrdClient, error) {

	if crdClient != nil {
		return crdClient, nil
	}

	kubecfg := path.Join(os.Getenv("HOME"), kubecfgsfx)
	cfg, err := clientcmd.BuildConfigFromFlags(master, kubecfg)
	Expect(err).NotTo(HaveOccurred())

	crdClient, err := crd.NewCrdClient(cfg, nil)
	Expect(err).NotTo(HaveOccurred())

	err = crdClient.CreateCRDs()
	Expect(err).NotTo(HaveOccurred())

	return crdClient, nil
}
