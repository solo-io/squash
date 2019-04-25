package e2e_test

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/squash/test/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

/* In order for these tests to run, the following env vars must be set:
PLANK_IMAGE_TAG
PLANK_IMAGE_REPO
*/
func TestE2e(t *testing.T) {

	helpers.RegisterCommonFailHandlers()
	helpers.SetupLog()

	RunSpecs(t, "E2e Squash Suite")
}

var testConditions = testutils.TestConditions{}

var _ = BeforeSuite(func() {
	err := testutils.InitializeTestConditions(&testConditions)
	Expect(err).NotTo(HaveOccurred())
	fmt.Println(testutils.SummarizeTestConditions(testConditions))

	seed := time.Now().UnixNano()
	fmt.Printf("rand seed: %v\n", seed)
	rand.Seed(seed)
})

// this list will be append each time a test namespace is created
var squashTestNamespaces = []string{}
var _ = AfterSuite(func() {
	fmt.Println("clean up after test")
	cs := MustGetClientset()
	for _, ns := range squashTestNamespaces {
		if err := cs.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{}); err != nil {
			// don't fail if cleanup fails
			fmt.Printf("Could not delete namespace %v, %v", ns, err)
		}
	}
})
