package e2e_test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sqtestutils "github.com/solo-io/squash/test/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

func TestE2e(t *testing.T) {
	testutils.RegisterPreFailHandler(
		func() {
			testutils.PrintTrimmedStack()
		})
	testutils.RegisterCommonFailHandlers()

	RunSpecs(t, "E2e Squash Suite")
}

const pathToBuildSpec = "../../solo-project.yaml"

var testConditions = sqtestutils.TestConditions{}

var _ = BeforeSuite(func() {
	err := sqtestutils.InitializeTestConditions(&testConditions, pathToBuildSpec)
	Expect(err).NotTo(HaveOccurred())
	fmt.Println(sqtestutils.SummarizeTestConditions(testConditions))

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
