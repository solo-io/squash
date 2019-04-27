package e2e_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/squash/pkg/config"
	sqOpts "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/test/testutils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("Single debug mode", func() {

	It("Should create a debug session", func() {
		By("should get a kube client")
		cs := MustGetClientset()

		By("should list no resources after delete")
		// Run delete before testing to ensure there are no lingering artifacts
		must(testutils.Squashctl("utils delete-attachments"))
		str, err := testutils.SquashctlOut("utils list-attachments")
		check(err)
		validateUtilsListDebugAttachments(str, 0)

		// create namespace
		testNamespace := fmt.Sprintf("testsquash-demos-%v", rand.Intn(1000))
		testPlankNamespace := fmt.Sprintf("testsquash-planks-%v", rand.Intn(1000))
		testPlankNamespace = sqOpts.SquashNamespace // TODO(mitchdraft) - unhardcode this when plank reads from os.Env
		By("should create a demo namespace")
		_, err = cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}})
		check(err)
		_, _ = cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testPlankNamespace}})
		// TODO(mitchdraft) - use below format when SquashNamespace is generalized
		//_, err = cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testPlankNamespace}})
		//check(err)
		squashTestNamespaces = append(squashTestNamespaces, testNamespace)
		squashTestNamespaces = append(squashTestNamespaces, testPlankNamespace)

		By("should deploy a demo app")
		must(testutils.Squashctl(fmt.Sprintf("deploy demo --demo-id %v --demo-namespace1 %v --demo-namespace2 %v", "go-java",
			testNamespace,
			testPlankNamespace)))

		By("should find the demo deployment")
		goPodName, err := waitForPod(cs, testNamespace, "example-service1")
		check(err)
		javaPodName, err := waitForPod(cs, testPlankNamespace, "example-service2-java")
		check(err)

		By("should attach a dlv debugger")
		dbgStr, err := testutils.SquashctlOut(testutils.MachineDebugArgs(testConditions, "dlv", testNamespace, goPodName, testPlankNamespace))
		check(err)
		validateMachineDebugOutput(dbgStr)

		By("should have created the required permissions")
		must(ensurePlankPermissionsWereCreated(cs, testPlankNamespace))

		By("should speak with dlv")
		ensureDLVServerIsLive(dbgStr)

		By("should list expected resources after debug session initiated")
		attachmentList, err := testutils.SquashctlOut("utils list-attachments")
		check(err)
		validateUtilsListDebugAttachments(attachmentList, 1)

		By("utils delete-planks should not delete non-plank pods")
		plankNsPods := mustGetActivePlankNsPods(cs, testPlankNamespace)
		// should be one plank and one java demo service
		Expect(len(plankNsPods)).To(Equal(2))
		podsMustInclude(plankNsPods, javaPodName)
		must(testutils.Squashctl(fmt.Sprintf("utils delete-planks")))
		plankNsPods = mustGetActivePlankNsPods(cs, testPlankNamespace)
		ExpectWithOffset(1, len(plankNsPods)).To(Equal(1))
		ExpectWithOffset(1, plankNsPods[0].Name).To(Equal(javaPodName))

		// cleanup
		By("should cleanup")
		check(cs.CoreV1().Namespaces().Delete(testNamespace, &metav1.DeleteOptions{}))
	})
	It("should delete planks only", func() {

	})
})

func waitForPod(cs *kubernetes.Clientset, testNamespace, deploymentName string) (string, error) {
	// this can be slow, pulls image for the first time - should store demo images in cache if possible
	timeLimit := 100
	timeStepSleepDuration := time.Second
	for i := 0; i < timeLimit; i++ {
		pods, err := cs.CoreV1().Pods(testNamespace).List(metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		if podName, found := findPod(pods, deploymentName); found {
			return podName, nil
		}
		time.Sleep(timeStepSleepDuration)
	}
	return "", fmt.Errorf("Pod for deployment %v not found", deploymentName)
}

func findPod(pods *v1.PodList, deploymentName string) (string, bool) {
	for _, pod := range pods.Items {
		if pod.Spec.Containers[0].Name == deploymentName && podReady(pod) {
			return pod.Name, true
		}
	}
	return "", false
}

func podReady(pod v1.Pod) bool {
	switch pod.Status.Phase {
	case v1.PodRunning:
		return true
	case v1.PodSucceeded:
		return true
	default:
		return false
	}
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
	ExpectWithOffset(1, lines[0]).To(Equal(expectedHeader))
	ExpectWithOffset(1, len(lines)).To(Equal(expectedLength))
	for i := 1; i < expectedLength; i++ {
		validateUtilsListDebugAttachmentsLine(lines[i])
	}
}

func validateUtilsListDebugAttachmentsLine(line string) {
	cols := strings.Split(line, ", ")
	ExpectWithOffset(3, len(cols)).To(Equal(2))
}

/* sample of expected output:
{"PortForwardCmd":"kubectl port-forward plankhxpq4 :33303 -n squash-debugger"}
*/
func validateMachineDebugOutput(output string) {
	re := regexp.MustCompile(`{"PortForwardCmd":"kubectl port-forward.*}`)
	By(output)
	ExpectWithOffset(1, re.MatchString(output)).To(BeTrue())
}

// using the kubectl port-forward command spec provided by the Plank pod,
// port forward, curl, and inspect the curl error message
// expect to see the error associated with a rejection, rather than a failure to connect
func ensureDLVServerIsLive(dbgJson string) {
	ed := config.EditorData{}
	check(json.Unmarshal([]byte(dbgJson), &ed))
	cmdParts := strings.Split(ed.PortForwardCmd, " ")
	// 0: kubectl
	// 1:	port-forward
	// 2: plankhxpq4
	// 3: :33303
	// 4: -n
	// 5: squash-debugger
	ports := strings.Split(cmdParts[3], ":")
	remotePort := ports[1]
	var localPort int
	check(utils.FindAnyFreePort(&localPort))
	cmdParts[3] = fmt.Sprintf("%v:%v", localPort, remotePort)
	// the portforward spec includes "kubectl ..." but exec.Command requires the binary be called explicitly
	pfCmd := exec.Command("kubectl", cmdParts[1:]...)
	go func() {
		out, _ := pfCmd.CombinedOutput()
		fmt.Println(string(out))
	}()
	time.Sleep(2 * time.Second)
	dlvAddr := fmt.Sprintf("localhost:%v", localPort)
	curlOut, _ := testutils.Curl(dlvAddr)
	// valid response signature: curl: (52) Empty reply from server
	// invalid response signature: curl: (7) Failed to connect to localhost port 58239: Connection refused
	re := regexp.MustCompile(`curl: \(52\) Empty reply from server`)
	match := re.Match(curlOut)
	Expect(match).To(BeTrue())
	// dlvClient := rpc1.NewClient(dlvAddr)
	// err, dlvState := dlvClient.GetState()
	// check(err)
}

func ensurePlankPermissionsWereCreated(cs *kubernetes.Clientset, plankNs string) error {
	if _, err := cs.CoreV1().ServiceAccounts(plankNs).Get(sqOpts.PlankServiceAccountName, metav1.GetOptions{}); err != nil {
		return err
	}
	if _, err := cs.RbacV1().ClusterRoles().Get(sqOpts.PlankClusterRoleName, metav1.GetOptions{}); err != nil {
		return err
	}
	if _, err := cs.RbacV1().ClusterRoleBindings().Get(sqOpts.PlankClusterRoleBindingName, metav1.GetOptions{}); err != nil {
		return err
	}
	return nil
}
func mustGetActivePlankNsPods(cs *kubernetes.Clientset, plankNs string) []v1.Pod {
	allPods, err := cs.CoreV1().Pods(plankNs).List(metav1.ListOptions{})
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	ExpectWithOffset(1, allPods).NotTo(BeNil())
	pods := []v1.Pod{}
	for _, p := range allPods.Items {
		if p.ObjectMeta.DeletionTimestamp == nil {
			pods = append(pods, p)
		}
	}
	return pods
}

func podsMustInclude(pods []v1.Pod, name string) {
	foundPod := false
	for _, p := range pods {
		if p.Name == name {
			foundPod = true
		}
	}
	ExpectWithOffset(1, foundPod).To(BeTrue())
}
