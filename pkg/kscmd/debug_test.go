package kscmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/solo-io/squash/pkg/lite/kube"
)

var _ = Describe("Debug", func() {
	It("should get image from yaml", func() {
		image, _, err := SkaffoldConfigToPod("https://raw.githubusercontent.com/GoogleContainerTools/skaffold/master/examples/getting-started/skaffold.yaml")
		Expect(err).To(Not(HaveOccurred()))
		Expect(image).To(Equal("gcr.io/k8s-skaffold/skaffold-example"))
		// Expect(podname).To(Equal("getting-started"))
	})
})
