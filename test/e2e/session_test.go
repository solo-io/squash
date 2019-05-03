package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha2"

	skube "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	squashv1 "github.com/solo-io/squash/pkg/api/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/solo-io/squash/pkg/config"
	sqOpts "github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"
	"github.com/solo-io/squash/test/testutils"
	v1 "k8s.io/api/core/v1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("Single debug mode", func() {
	const namespaceOfSeedImagePullSecret = "default"
	var (
		testNamespace      string
		testPlankNamespace string
		cs                 *kubernetes.Clientset
		logCtx             context.Context
		cancelFunc         context.CancelFunc
	)

	BeforeEach(func() {
		cs = MustGetClientset()
		logCtx, cancelFunc = context.WithCancel(context.Background())
		go func() {
			if err := dumpLogsBackground(logCtx, true); err != nil {
				log.Fatal(err)
			}
		}()

		testNamespace = fmt.Sprintf("testsquash-demos-%v", rand.Intn(1000))
		testPlankNamespace = fmt.Sprintf("testsquash-planks-%v", rand.Intn(1000))
		By("should create a demo namespace")
		_, err := cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}})
		Expect(err).NotTo(HaveOccurred())
		_, err = cs.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testPlankNamespace}})
		Expect(err).NotTo(HaveOccurred())
		squashTestNamespaces = append(squashTestNamespaces, testNamespace)
		squashTestNamespaces = append(squashTestNamespaces, testPlankNamespace)
		applyPullSecretsMessage, err := applyImagePullSecretsToDefaultServiceAccount(cs, namespaceOfSeedImagePullSecret, sqOpts.SquashServiceAccountImagePullSecretName, []string{testNamespace, testPlankNamespace})
		Expect(err).NotTo(HaveOccurred())
		By(applyPullSecretsMessage)

		// Run delete before testing to ensure there are no lingering artifacts
		By("should list no resources after delete")
		_, err = testutils.SquashctlOut("utils delete-attachments")
		Expect(err).NotTo(HaveOccurred())
		str, err := testutils.SquashctlOut("utils list-attachments")
		Expect(err).NotTo(HaveOccurred())

		validateUtilsListDebugAttachments(str, 0)
	})

	AfterEach(func() {
		unregisterDebugAttachmentCRD()
		cs := MustGetClientset()
		err := cs.CoreV1().Namespaces().Delete(testNamespace, &metav1.DeleteOptions{})
		Expect(err).NotTo(HaveOccurred())
		// need to delete permissions after each test, not only for good citizenship, but because squash only supports
		// one instance of the permission set right now, and if a new plank namespace is used (as in these tests), the
		// old permission set will not work as expected
		err = testutils.Squashctl("utils delete-permissions")
		Expect(err).NotTo(HaveOccurred())
		cancelFunc()
	})

	It("Should create a debug session", func() {
		goPodName, javaPodName := installSquashBuiltInDemoApps(cs, testNamespace, testPlankNamespace)
		By("should attach a dlv debugger")
		dbgStr, err := testutils.SquashctlOut(testutils.MachineDebugArgs(testConditions, "dlv", testNamespace, goPodName, testPlankNamespace, "", ""))
		Expect(err).NotTo(HaveOccurred())

		By("should have created the required permissions")
		err = ensurePlankPermissionsWereCreated(cs, testPlankNamespace)
		Expect(err).NotTo(HaveOccurred())
		validateMachineDebugOutput(dbgStr)

		By("should speak with dlv")
		ensureDLVServerIsLive(dbgStr)

		By("should list expected resources after debug session initiated")
		attachmentList, err := testutils.SquashctlOut("utils list-attachments")
		Expect(err).NotTo(HaveOccurred())
		validateUtilsListDebugAttachments(attachmentList, 1)

		By("utils delete-planks should not delete non-plank pods")
		plankNsPods := mustGetActivePlankNsPods(cs, testPlankNamespace)
		// should be one plank and one java demo service
		Expect(len(plankNsPods)).To(Equal(2))
		podsMustInclude(plankNsPods, javaPodName)
		err = testutils.Squashctl(fmt.Sprintf("utils delete-planks --squash-namespace %v", testPlankNamespace))
		Expect(err).NotTo(HaveOccurred())
		plankNsPods = mustGetActivePlankNsPods(cs, testPlankNamespace)
		Expect(len(plankNsPods)).To(Equal(1))
		Expect(plankNsPods[0].Name).To(Equal(javaPodName))
	})

	It("Should create a debug session - secure mode", func() {
		goPodName, javaPodName := installSquashBuiltInDemoApps(cs, testNamespace, testPlankNamespace)
		configFile := "secure_mode_test_config.yaml"
		err := testutils.Squashctl(fmt.Sprintf("deploy squash --squash-namespace %v --container-repo %v --container-version %v --config %v",
			testPlankNamespace,
			testConditions.PlankImageRepo,
			testConditions.PlankImageTag,
			configFile,
		))
		Expect(err).NotTo(HaveOccurred())

		By("should have created a squash pod")
		squashPodName, err := waitForPod(cs, testPlankNamespace, sqOpts.SquashPodName)
		fmt.Println(squashPodName)
		Expect(err).NotTo(HaveOccurred())

		By("should attach a dlv debugger")
		dbgStr, err := testutils.SquashctlOut(testutils.MachineDebugArgs(testConditions, "dlv", testNamespace, goPodName, testPlankNamespace, configFile, ""))
		Expect(err).NotTo(HaveOccurred())
		validateMachineDebugOutput(dbgStr)

		By("should have created the required permissions")
		err = ensurePlankPermissionsWereCreated(cs, testPlankNamespace)
		Expect(err).NotTo(HaveOccurred())

		By("should speak with dlv")
		ensureDLVServerIsLive(dbgStr)

		By("should list expected resources after debug session initiated")
		attachmentList, err := testutils.SquashctlOut("utils list-attachments")
		Expect(err).NotTo(HaveOccurred())
		validateUtilsListDebugAttachments(attachmentList, 1)

		By("utils delete-planks should not delete non-plank pods")
		plankNsPods := mustGetActivePlankNsPods(cs, testPlankNamespace)
		// should be one squash, one plank, and one java demo service
		Expect(len(plankNsPods)).To(Equal(3))
		podsMustInclude(plankNsPods, javaPodName)
		err = testutils.Squashctl(fmt.Sprintf("utils delete-planks --squash-namespace %v", testPlankNamespace))
		Expect(err).NotTo(HaveOccurred())
		plankNsPods = mustGetActivePlankNsPods(cs, testPlankNamespace)
		// should include one squash and one java demo service
		Expect(len(plankNsPods)).To(Equal(2))
		Expect(plankNsPods[0].Name).To(Equal(javaPodName))
	})

	It("Should debug specific process - default, single-process case", func() {
		multiProcessTest(cs, testNamespace, testPlankNamespace, "", true)
	})

	It("Should debug specific process - multi-process case", func() {
		processName := "sample_app"
		multiProcessTest(cs, testNamespace, testPlankNamespace, processName, false)
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

func waitForPodByLabel(cs *kubernetes.Clientset, testNamespace, labelSelector string) (string, error) {
	// this can be slow, pulls image for the first time - should store demo images in cache if possible
	timeLimit := 100
	timeStepSleepDuration := time.Second
	for i := 0; i < timeLimit; i++ {
		pods, err := cs.CoreV1().Pods(testNamespace).List(metav1.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			return "", err
		}
		if len(pods.Items) == 1 {
			return pods.Items[0].Name, nil
		}
		if len(pods.Items) > 1 {
			return "", fmt.Errorf("label selector %v returned %v matches (one expected)", labelSelector, len(pods.Items))
		}
		time.Sleep(timeStepSleepDuration)
	}
	return "", fmt.Errorf("Pod with label selector %v not found", labelSelector)
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
	By(fmt.Sprintf("Output from validateMachineDebugOutput: %v", output))
	ExpectWithOffset(1, re.MatchString(output)).To(BeTrue())
}

// using the kubectl port-forward command spec provided by the Plank pod,
// port forward, curl, and inspect the curl error message
// expect to see the error associated with a rejection, rather than a failure to connect
func ensureDLVServerIsLive(dbgJson string) {
	ed := config.EditorData{}
	err := json.Unmarshal([]byte(dbgJson), &ed)
	Expect(err).NotTo(HaveOccurred())
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
	err = utils.FindAnyFreePort(&localPort)
	Expect(err).NotTo(HaveOccurred())
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
	By(fmt.Sprintf("looking for pod name: %v", name))
	Expect(foundPod).To(BeTrue())
}

// This util does two things, you may only need one of them:
// 1. Copies the specified secret to the specified namespaces
// 2. Grants the secret as an ImagePullSecret to each of the default service accounts in the specified namespaces
// If your pod specifies an image pull secret, you need (1)
// If the namespace's default service account is pulling the images, you need (2)
func applyImagePullSecretsToDefaultServiceAccount(cs *kubernetes.Clientset, fromNs, secretName string, toNamespaces []string) (string, error) {
	imagePullSecret, err := cs.CoreV1().Secrets(fromNs).Get(secretName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return "NOT applying image pull secrets", nil
	}
	Expect(err).NotTo(HaveOccurred())
	for _, ns := range toNamespaces {
		secretRef := core.ResourceRef{Namespace: ns, Name: sqOpts.SquashServiceAccountImagePullSecretName}
		copyImagePullSecretAndGrantToServiceAccount(cs, imagePullSecret, secretRef, "default")
	}
	// make sure that the service account updates before the image pull begins
	time.Sleep(200 * time.Millisecond)
	return fmt.Sprintf("Applied %v image pull secrets", len(toNamespaces)), nil
}
func copyImagePullSecretAndGrantToServiceAccount(cs *kubernetes.Clientset, imagePullSecret *v1.Secret, toSecretRef core.ResourceRef, toServiceAccountName string) {
	imagePullSecret.ObjectMeta.Namespace = toSecretRef.Namespace
	_, err := cs.CoreV1().Secrets(toSecretRef.Namespace).Create(&v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imagePullSecret.ObjectMeta.Name,
			Namespace: toSecretRef.Namespace,
		},
		Data: imagePullSecret.Data,
		Type: imagePullSecret.Type,
	})
	if errors.IsAlreadyExists(err) {
		By(fmt.Sprintf("secret %v already exists, bailing from copyImagePullSecretAndGrantToServiceAccount", toSecretRef.Name))
		return
	}
	Expect(err).NotTo(HaveOccurred())
	sa, err := cs.CoreV1().ServiceAccounts(toSecretRef.Namespace).Get(toServiceAccountName, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	sa.ImagePullSecrets = append(sa.ImagePullSecrets, v1.LocalObjectReference{Name: sqOpts.SquashServiceAccountImagePullSecretName})
	_, err = cs.CoreV1().ServiceAccounts(toSecretRef.Namespace).Update(sa)
	// retry once if there's an error
	if errors.IsConflict(err) {
		time.Sleep(200 * time.Millisecond)
		sa, err := cs.CoreV1().ServiceAccounts(toSecretRef.Namespace).Get(toServiceAccountName, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, v1.LocalObjectReference{Name: sqOpts.SquashServiceAccountImagePullSecretName})
		_, err = cs.CoreV1().ServiceAccounts(toSecretRef.Namespace).Update(sa)
		Expect(err).NotTo(HaveOccurred())
	}
}

func unregisterDebugAttachmentCRD() {
	cfg, err := kubeutils.GetConfig("", "")
	Expect(err).NotTo(HaveOccurred())
	extClient, err := apiexts.NewForConfig(cfg)
	err = extClient.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(squashv1.DebugAttachmentCrd.FullName(), &metav1.DeleteOptions{})
	if !errors.IsNotFound(err) {
		Expect(err).NotTo(HaveOccurred())
	}
}

// very basic use of skaffold's log aggregation utilities: dump everything
// will polish this later with a go-util
type podFilter struct {
	name string
}

func (pf podFilter) Select(pod *v1.Pod) bool {
	return true // for now anyway
	//if pod.Name == pf.name {
	//	return true
	//}
	//return false
}

type color int

func (pf podFilter) Pick(pod *v1.Pod) color {
	// just return red for now
	return 31
}

// you should call cancel on your context when done with the logs
func dumpLogsBackground(ctx context.Context, onFailOnly bool) error {
	// these are mostly placeholders
	ps := &podFilter{name: "tmp"}
	arts := []*v1alpha2.Artifact{{
		ImageName: "",
		Workspace: "",
	}}
	cp := skube.NewColorPicker(arts)
	la := &skube.LogAggregator{}
	if onFailOnly {
		la = skube.NewLogAggregator(GinkgoWriter, ps, cp)
	} else {
		la = skube.NewLogAggregator(os.Stdout, ps, cp)
	}
	if err := la.Start(ctx); err != nil {
		return err
	}
	return nil
}

//func applyManifest(filepath, ns string) {
//	_, err := helmchart.RenderManifests(context.Background(), filepath, "", "", ns, "")
//	Expect(err).NotTo(HaveOccurred())
//}

func installSquashBuiltInDemoApps(cs *kubernetes.Clientset, appNamespace, plankNamespace string) (string, string) {
	By("should deploy a demo app")
	err := testutils.Squashctl(fmt.Sprintf("deploy demo --demo-id %v --demo-namespace1 %v --demo-namespace2 %v", "go-java",
		appNamespace,
		plankNamespace))
	Expect(err).NotTo(HaveOccurred())

	By("should find the demo deployment")
	goPodName, err := waitForPod(cs, appNamespace, "example-service1")
	Expect(err).NotTo(HaveOccurred())
	javaPodName, err := waitForPod(cs, plankNamespace, "example-service2-java")
	Expect(err).NotTo(HaveOccurred())
	return goPodName, javaPodName
}
