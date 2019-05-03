package e2e_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gotestutils "github.com/solo-io/go-utils/testutils"
	"github.com/solo-io/squash/test/testutils"
	"k8s.io/client-go/kubernetes"
)

// To verify that this test does what you expect on a simple container, set verifySelfMode=false
// Suggestion: run this function twice in adjacent It() blocks, use the same args, except set verifySelfMode=true for
// the first call, and false for the second call
func multiProcessTest(cs *kubernetes.Clientset, appNamespace, plankNamespace string, verifySelfMode bool) {
	installFile := "../../contrib/condition/multi_process/multi.yaml"
	labelSelector := "app=squash-demo-multiprocess"
	if verifySelfMode {
		installFile = "../../contrib/condition/multi_process/single.yaml"
		labelSelector = "app=squash-demo-multiprocess-base"
	}
	appOut, err := gotestutils.KubectlOut("apply", "-f", installFile, "-n", appNamespace)
	fmt.Fprintf(GinkgoWriter, appOut)
	Expect(err).NotTo(HaveOccurred())
	//applyManifest("../../../contrib/condition/multi_process/single.yaml", testNamespace)
	appName, err := waitForPodByLabel(cs, appNamespace, labelSelector)
	Expect(err).NotTo(HaveOccurred())
	By("should attach a dlv debugger")

	By("starting debug session")
	timeLimitSeconds := 10
	dbgStr, err := testutils.SquashctlOutWithTimeout(testutils.MachineDebugArgs(testConditions,
		"dlv",
		appNamespace,
		appName,
		plankNamespace,
		""), &timeLimitSeconds)
	Expect(err).NotTo(HaveOccurred())

	By("should have created the required permissions")
	err = ensurePlankPermissionsWereCreated(cs, plankNamespace)
	Expect(err).NotTo(HaveOccurred())
	validateMachineDebugOutput(dbgStr)

	By("should speak with dlv")
	ensureDLVServerIsLive(dbgStr)

	By("should list expected resources after debug session initiated")
	attachmentList, err := testutils.SquashctlOut("utils list-attachments")
	Expect(err).NotTo(HaveOccurred())
	validateUtilsListDebugAttachments(attachmentList, 1)

	By("utils delete-planks should not delete non-plank pods")
	err = testutils.Squashctl(fmt.Sprintf("utils delete-planks --squash-namespace %v", plankNamespace))
	Expect(err).NotTo(HaveOccurred())
	appPods := mustGetActivePlankNsPods(cs, appNamespace)
	Expect(len(appPods)).To(Equal(1))
	Expect(appPods[0].Name).To(Equal(appName))
}
