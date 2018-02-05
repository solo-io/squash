package crd_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/platforms/kubernetes/crd"
	k8models "github.com/solo-io/squash/pkg/platforms/kubernetes/models"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	kubecfg    = "/Users/anton/.kube/config"
	master     = ""
	dbgName    = "testDebugger"
	badDbgName = "badDebugger"
	numInList  = 10
)

var crdClient *crd.CrdClient

func cb(objtype crd.ObjectType, action crd.Action, obj interface{}) {
	fmt.Println("Callback:", objtype, action)
	if objtype == crd.TypeDebugAttachment {
		v := obj.(*models.DebugAttachment)
		fmt.Println("DebugAttachment: ", v.Metadata.Name)
	} else {
		v := obj.(*models.DebugRequest)
		fmt.Println("DebugRequest: ", v.Metadata.Name)
	}
}

func init() {
	var err error
	crdClient, err = initTests(cb)
	if err != nil {
		fmt.Errorf("Test initialization failed: %v", err)
	}
}

func TestAttachment(t *testing.T) {
	if crdClient == nil {
		return
	}

	a, err := createAttachment()
	if err != nil {
		t.Fatalf("Error creating attachment: %s", err.Error())
	}

	a.Spec.Debugger = dbgName
	a, err = crdClient.UpdateAttachment(a)
	if err != nil {
		t.Fatalf("Error updating attachment: %s", err.Error())
	}

	ab, err := crdClient.GetAttachment(a.Metadata.Name)
	if err != nil {
		t.Fatalf("Error getting attachment: %s", err.Error())
	}

	if a.Spec.Debugger != ab.Spec.Debugger {
		t.Fatalf("Saved and retrived attachments are not identical: %s != %s", a.Spec.Debugger, ab.Spec.Debugger)
	}

	ka := a.Spec.Attachment.(*k8models.KubeAttachment)

	t.Log(ka)

	err = crdClient.DeleteAttachment(a.Metadata.Name)
	if err != nil {
		t.Fatalf("Error deleting attachment: %s", err.Error())
	}
}

func TestRequest(t *testing.T) {
	if crdClient == nil {
		return
	}

	r, err := createRequest()
	if err != nil {
		t.Fatalf("Error creating request: %s", err.Error())
	}

	dbgNameGood := dbgName
	r.Spec.Debugger = &dbgNameGood

	r, err = crdClient.UpdateRequest(r)
	if err != nil {
		t.Fatalf("Error updating request: %s", err.Error())
	}

	rb, err := crdClient.GetRequest(r.Metadata.Name)
	if err != nil {
		t.Fatalf("Error getting request: %s", err.Error())
	}

	if *r.Spec.Debugger != *rb.Spec.Debugger {
		t.Fatalf("Saved and retrived requests are not identical: %s != %s", r.Spec.Debugger, rb.Spec.Debugger)
	}

	err = crdClient.DeleteRequest(r.Metadata.Name)
	if err != nil {
		t.Fatalf("Error deleting request: %s", err.Error())
	}
}

func TestAttachmentsList(t *testing.T) {
	if crdClient == nil {
		return
	}

	for i := 0; i < numInList; i++ {
		_, err := createAttachment()
		if err != nil {
			t.Fatalf("Error creating attachment: %s", err.Error())
		}
	}

	v, err := crdClient.ListAttachments()
	if err != nil {
		t.Fatalf("Error listing attachments: %s", err.Error())
	}
	if len(v) != numInList {
		t.Fatalf("Wrong number of items in list: %d (expected %d)", len(v), numInList)
	}

	for _, i := range v {
		t.Logf("Item: %s", i.Metadata.Name)
		err = crdClient.DeleteAttachment(i.Metadata.Name)
		if err != nil {
			t.Errorf("Error deleting attachment: %s", err.Error())
		}
	}
}

func TestRequestsList(t *testing.T) {
	if crdClient == nil {
		return
	}

	for i := 0; i < numInList; i++ {
		_, err := createRequest()
		if err != nil {
			t.Fatalf("Error creating request: %s", err.Error())
		}
	}

	v, err := crdClient.ListRequests()
	if err != nil {
		t.Fatalf("Error listing requests: %s", err.Error())
	}
	if len(v) != numInList {
		t.Fatalf("Wrong number of items in list: %d (expected %d)", len(v), numInList)
	}

	for _, i := range v {
		t.Logf("Item: %s", i.Metadata.Name)
		err = crdClient.DeleteRequest(i.Metadata.Name)
		if err != nil {
			t.Errorf("Error deleting request: %s", err.Error())
		}
	}
}

func TestCleanup(t *testing.T) {
	cleanupAttachments()
	cleanupRequests()
}

func TestUpdate(t *testing.T) {
	//	t.Skip("Debug only")
	if crdClient == nil {
		return
	}

	a, err := createAttachment()
	if err != nil {
		t.Fatalf("Error creating attachment: %s", err.Error())
	}

	time.Sleep(290 * time.Second)
	crdClient.DeleteAttachment(a.Metadata.Name)
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

func cleanupAttachments() error {
	v, err := crdClient.ListAttachments()
	if err != nil {
		return err
	}

	for _, i := range v {
		crdClient.DeleteAttachment(i.Metadata.Name)
	}
	return nil
}

func cleanupRequests() error {
	v, err := crdClient.ListRequests()
	if err != nil {
		return err
	}

	for _, i := range v {
		crdClient.DeleteRequest(i.Metadata.Name)
	}
	return nil
}

func initTests(cb func(objtype crd.ObjectType, action crd.Action, obj interface{})) (*crd.CrdClient, error) {

	if crdClient != nil {
		return crdClient, nil
	}

	cfg, err := clientcmd.BuildConfigFromFlags(master, kubecfg)
	if err != nil {
		return nil, fmt.Errorf("Error building kubeconfig: %s", err.Error())
	}

	crdClient, err := crd.NewCrdClient(cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating client: %s", err.Error())
	}

	err = crdClient.CreateCRDs()
	if err != nil {
		return nil, fmt.Errorf("Error creating CRDs attachment: %s", err.Error())
	}
	return crdClient, nil
}
