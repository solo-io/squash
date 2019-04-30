package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"k8s.io/client-go/kubernetes"
)

func MustGetClientset() *kubernetes.Clientset {
	cs := &kubernetes.Clientset{}
	By("should get a kube client")
	restCfg, err := gokubeutils.GetConfig("", "")
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	cs, err = kubernetes.NewForConfig(restCfg)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return cs
}
