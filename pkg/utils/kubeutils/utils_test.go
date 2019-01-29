package kubeutils_test

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/squash/pkg/utils/kubeutils"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("Get namespaces", func() {
	It("should get namespaces", func() {
		restCfg, err := gokubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		cs, err := kubernetes.NewForConfig(restCfg)
		Expect(err).NotTo(HaveOccurred())

		// create known ns
		name := fmt.Sprintf("ns%v", rand.Int63n(100000))
		newNs := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
		cs.CoreV1().Namespaces().Create(newNs)

		// test function
		ns := kubeutils.MustGetNamespaces(nil)
		Expect(len(ns) > 0).To(BeTrue())
		Expect(ns).To(ContainElement(name))

		// cleanup
		cs.CoreV1().Namespaces().Delete(name, nil)
	})
})
