package e2e_test

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gokubeutils "github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/squash/test/testutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func check(err error) {
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
}

var _ = Describe("Single debug mode", func() {
	testutils.DeclareTestConditions()

	seed := time.Now().UnixNano()
	fmt.Printf("rand seed: %v\n", seed)
	rand.Seed(seed)

	It("Should create a debug session", func() {
		cs := &kubernetes.Clientset{}
		By("should get a kube client")
		restCfg, err := gokubeutils.GetConfig("", "")
		check(err)
		cs, err = kubernetes.NewForConfig(restCfg)
		check(err)

		By("should list no resources after delete")
		err = testutils.Squashctl("utils delete-attachments")
		check(err)
		str, err := testutils.SquashctlOut("utils list-attachments")
		check(err)
		validateUtilsListDebugAttachments(str, 0)

		// create namespace
		testNamespace := fmt.Sprintf("testsquash-%v", rand.Intn(1000))
		By("should create a demo namespace")
		_, err = cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}})
		check(err)

		By("should deploy a demo app")
		err = testutils.Squashctl(fmt.Sprintf("deploy demo --demo-id %v --demo-namespace1 %v --demo-namespace2 %v", "go-java", testNamespace, testNamespace))
		check(err)
		time.Sleep(5 * time.Second)

		By("should find the demo deployment")
		pods, err := cs.CoreV1().Pods(testNamespace).List(metav1.ListOptions{})
		podName, err := findPod(pods, "example-service1")
		check(err)

		By("should attach a debugger")
		dbgStr, err := testutils.SquashctlOut(testutils.MachineDebugArgs("dlv", testNamespace, podName))
		check(err)
		// TODO(mitchdraft) - this is failing because the image needs to be pulled, add wait logic and re-enable
		// validateMachineDebugOutput(dbgStr)
		fmt.Println(dbgStr)

		By("should list expected resources after debug session initiated")
		attachmentList, err := testutils.SquashctlOut("utils list-attachments")
		check(err)
		validateUtilsListDebugAttachments(attachmentList, 1)

		// cleanup
		By("should cleanup")
		check(cs.CoreV1().Namespaces().Delete(testNamespace, &metav1.DeleteOptions{}))
	})
})

func findPod(pods *v1.PodList, deploymentName string) (string, error) {
	for _, pod := range pods.Items {
		if pod.Spec.Containers[0].Name == deploymentName {
			return pod.Name, nil
		}
	}
	return "", fmt.Errorf("Pod for deployment %v not found", deploymentName)
}

/* sample of expected output (case of 4 debug attachments across two namespaces)
Existing debug attachments:
dd, ea8f2f3omi
dd, hm52rfvkbt
default, cq13qxkxa2
default, lmgv6h2g7o
*/
func validateUtilsListDebugAttachments(output string, expectedDaCount int) {
	lines := strings.Split(output, "\n")
	// should return one line per da + a header line
	expectedLength := 1 + expectedDaCount
	expectedHeader := "Existing debug attachments:"
	if expectedDaCount == 0 {
		expectedHeader = "Found no debug attachments"
	}
	Expect(lines[0]).To(Equal(expectedHeader))
	Expect(len(lines)).To(Equal(expectedLength))
	for i := 1; i < expectedLength; i++ {
		validateUtilsListDebugAttachmentsLine(lines[i])
	}
}

func validateUtilsListDebugAttachmentsLine(line string) {
	cols := strings.Split(line, ", ")
	Expect(len(cols)).To(Equal(2))
}

/* sample of expected output:
{"PortForwardCmd":"kubectl port-forward plankhxpq4 :33303 -n squash-debugger"}
*/
func validateMachineDebugOutput(output string) {
	re := regexp.MustCompile(`{"PortForwardCmd":"kubectl port-forward.*}`)
	Expect(re.MatchString(output)).To(BeTrue())
}
