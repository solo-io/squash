package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	models "github.com/solo-io/squash/pkg/api/v1" // todo - rename this squash v1 after swagger extraction
	"github.com/solo-io/squash/pkg/options"
	"github.com/solo-io/squash/pkg/utils"

	// "github.com/solo-io/squash/pkg/models"

	v1 "k8s.io/api/core/v1"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Kubectl struct {
	Context, Namespace string

	proxyProcess *os.Process
	proxyAddress *string
}

func NewKubectl(kubectlctx string) *Kubectl {
	return &Kubectl{
		Context:   kubectlctx,
		Namespace: fmt.Sprintf("test-%d", rand.Uint64()),
	}
}

func (k *Kubectl) String() string {
	return fmt.Sprintf("context: %v, namespace: %v, proxyProcess: %v, proxyAddress: %v", k.Context, k.Namespace, *k.proxyProcess, *k.proxyAddress)
}

func (k *Kubectl) CreateNS() error {
	args := []string{"create", "namespace", k.Namespace}
	return k.innerprepare(args).Run()
}

func (k *Kubectl) DeleteNS() error {
	if k.Namespace != "" {
		args := []string{"delete", "namespace", k.Namespace}
		return k.innerprepare(args).Run()
	}
	return nil
}

func (k *Kubectl) CreateLocalRolesAndSleep(yamlfile string) error {
	return k.Apply(yamlfile, k.changedlocalrole, k.changenamespace, k.addsleep)
}

func (k *Kubectl) Create(yamlfile string) error {
	return k.Apply(yamlfile, k.changenamespace)
}

func (k *Kubectl) CreateSleep(yamlfile string) error {
	return k.Apply(yamlfile, k.changenamespace, k.addsleep)
}

func (k *Kubectl) changedlocalrole(yamlfile string) string {
	return strings.Replace(yamlfile, "ClusterRole", "Role", -1)
}

func (k *Kubectl) changenamespace(yamlfile string) string {
	return strings.Replace(yamlfile, "namespace: squash", "namespace: "+k.Namespace, -1)
}

func (k *Kubectl) addsleep(yamlfile string) string {
	// find image: soloio/squash-server
	// and add cmd and args
	regex := regexp.MustCompilePOSIX("^([[:space:]-]*)image: .*$")

	indexes := regex.FindStringSubmatchIndex(yamlfile)
	// first 1 indexes fir the match and second for the first (and only) submatch.
	if len(indexes) != 4 {
		panic("invalid yaml")
	}
	indentation := indexes[3] - indexes[2]
	insertion := indexes[1]
	indent := ""
	for i := 0; i < indentation; i++ {
		indent += " "
	}
	newyaml := yamlfile[:insertion] + "\n" + indent + "command: [\"sleep\", \"600\"]" + yamlfile[insertion:]

	return newyaml

}

func (k *Kubectl) Apply(yamlfile string, modifier ...func(string) string) error {

	b, err := ioutil.ReadFile(yamlfile)
	if err != nil {
		return err
	}
	yamlcontent := string(b)

	for _, m := range modifier {
		yamlcontent = m(yamlcontent)
	}

	buffer := bytes.NewBuffer(([]byte)(yamlcontent))

	cmd := k.innerpreparens([]string{"apply", "-f", "-"})
	cmd.Stdin = buffer

	return cmd.Run()

}

func (k *Kubectl) Pods() (*v1.PodList, error) {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	out, err := k.Prepare("get", "pods", "--output=json").Output()
	if err != nil {
		return nil, err
	}
	var pods v1.PodList
	err = json.Unmarshal(out, &pods)
	if err != nil {
		return nil, err
	}

	return &pods, nil
}

func (k *Kubectl) Exec(pod, container, cmd string, args ...string) error {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	prepareargs := []string{"exec", pod, "-c", container, "--", cmd}
	prepareargs = append(prepareargs, args...)
	return k.Prepare(prepareargs...).Run()
}

func (k *Kubectl) ExecAsync(pod, container, cmd string, args ...string) error {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	prepareargs := []string{"exec", pod, "-c", container, "--", cmd}
	prepareargs = append(prepareargs, args...)
	cmdtorun := k.Prepare(prepareargs...)

	err := cmdtorun.Start()
	if err == nil {
		go func() {
			cmdtorun.Wait()
		}()
	}
	return err
}

func (k *Kubectl) Cp(local, remote, pod, container string) error {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	return k.Prepare("cp", local, k.Namespace+"/"+pod+":"+remote, "-c", container).Run()
}

func (k *Kubectl) WaitPods(ctx context.Context) error {
OuterLoop:
	for {
		time.Sleep(5 * time.Second)
		select {
		case <-ctx.Done():
			return errors.New("timeout")
		default:
			pods, err := k.Pods()
			if err != nil {
				return err
			}
			for _, pod := range pods.Items {
				if pod.Status.Phase != v1.PodRunning {
					continue OuterLoop
				}
			}
			return nil
		}
	}
}

var proxyregex = regexp.MustCompile(`Starting to serve on\s+(\S+:\d+)`)

func (k *Kubectl) Proxy() error {
	cmd := exec.Command("kubectl", "proxy", "--port=0")

	portchan, err := runandreturn(cmd, proxyregex)
	if err != nil {
		return err
	}

	select {
	case port, ok := <-portchan:
		if ok {
			k.proxyProcess = cmd.Process
			k.proxyAddress = &port[1]
			return nil
		}
		cmd.Process.Kill()
		return errors.New("can't find port")
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		return errors.New("timeout")
	}
}
func (k *Kubectl) StopProxy() {
	if k.proxyProcess != nil {
		k.proxyProcess.Kill()
	}
}

var portregex = regexp.MustCompile(`from\s+\S+:(\d+)\s+->`)

func (k *Kubectl) PortForward(name string) (*os.Process, string, error) {
	// name is pod.namespace:port

	remoteparts := strings.Split(name, ":")
	if len(remoteparts) != 2 {
		return nil, "", errors.New("invalid remote")
	}
	podaddr := remoteparts[0]
	port := remoteparts[1]

	podparts := strings.Split(podaddr, ".")
	if len(podparts) != 2 {
		return nil, "", errors.New("invalid remote")
	}
	podName := podparts[0]
	podNamespace := podparts[1]
	args := []string{"--namespace=" + podNamespace, "port-forward", podName, ":" + port}
	cmd := k.innerprepare(args)

	portchan, err := runandreturn(cmd, portregex)
	if err != nil {
		return nil, "", err
	}

	select {
	case port, ok := <-portchan:
		if ok {
			return cmd.Process, "localhost:" + port[1], nil
		}
		cmd.Process.Kill()
		return nil, "", errors.New("can't find port")
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		return nil, "", errors.New("timeout")
	}
}

func runandreturn(cmd *exec.Cmd, reg *regexp.Regexp) (<-chan []string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Start()
	retchan := make(chan []string, 1)
	go func() {
		var buf bytes.Buffer
		for {
			b := make([]byte, 1024)
			n, err := stdout.Read(b)
			if err != nil || n == 0 {
				close(retchan)
				break
			}
			b = b[:n]
			buf.Write(b)

			std := buf.String()
			matches := reg.FindStringSubmatch(std)
			if len(matches) > 0 {
				retchan <- matches
				break
			}
		}
		cmd.Wait()
	}()
	return retchan, nil
}

func (k *Kubectl) Logs(name string) ([]byte, error) {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	return k.Prepare("logs", name).CombinedOutput()
}

func (k *Kubectl) Prepare(args ...string) *exec.Cmd {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	return k.innerpreparens(args)
}

func (k *Kubectl) innerunns(args []string) error {
	return k.innerpreparens(args).Run()
}
func (k *Kubectl) innerpreparens(args []string) *exec.Cmd {
	newargs := []string{"--namespace=" + k.Namespace}
	newargs = append(newargs, args...)
	return k.innerprepare(newargs)
}

func (k *Kubectl) innerprepare(args []string) *exec.Cmd {
	var newargs []string
	if k.Context != "" {
		newargs = []string{"--context=" + k.Context}
	}
	newargs = append(newargs, args...)
	fmt.Fprintln(GinkgoWriter, "kubectl", strings.Join(newargs, " "))
	cmd := exec.Command("kubectl", newargs...)
	return cmd
}

func NewSquash(k *Kubectl) *Squash {
	kubeaddr := "localhost:8001"
	if k.proxyAddress != nil {
		kubeaddr = *k.proxyAddress
	}
	return &Squash{
		Namespace: k.Namespace,
		kubeAddr:  kubeaddr,
	}
}

type Squash struct {
	Namespace string

	kubeAddr string
}

// Attach creates an attachment
func (s *Squash) Attach(image, pod, container, processName, dbgger string) (*models.DebugAttachment, error) {

	ctx := context.TODO() // TODO
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return nil, err
	}
	id := "name123"
	da := models.DebugAttachment{
		Metadata: core.Metadata{
			Name:      id,
			Namespace: options.SquashNamespace,
		},
		Debugger:  dbgger,
		Image:     image,
		Pod:       pod,
		Container: container,
		// DebugServerAddress: 	fmt.Sprintf("--url=http://"+s.kubeAddr+"/api/v1/namespaces/%s/services/squash-server:http-squash-api/proxy/api/v2", s.Namespace)
		DebugServerAddress: fmt.Sprintf("http://"+s.kubeAddr+"/api/v1/namespaces/%s/services/squash-server:http-squash-api/proxy/api/v2", s.Namespace),
	}
	// args := []string{"debug-container", "--namespace=" + s.Namespace, image, pod, container, dbgger}
	if processName != "" {
		// args = append(args, "--processName="+processName)
		da.ProcessName = processName
	}
	writeOpts := clients.WriteOpts{
		Ctx:               ctx,
		OverwriteExisting: false,
	}
	res, err := (*daClient).Write(&da, writeOpts)

	return res, err
}

func (s *Squash) Delete(da *models.DebugAttachment) error {
	args := []string{"delete", da.Metadata.Name}

	cmd := s.run(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		GinkgoWriter.Write(out)
		return err
	}

	return nil
}

func (s *Squash) Wait(id string) (*models.DebugAttachment, error) {

	cmd := s.run("wait", id, "--timeout", "90")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "Failed service wait:", string(out))
		return nil, err
	}

	var dbgattachment models.DebugAttachment
	err = json.Unmarshal(out, &dbgattachment)
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "Failed to parse response for service wait:", string(out))
		return nil, err
	}

	return &dbgattachment, nil
}

func (s *Squash) run(args ...string) *exec.Cmd {

	panic(strings.Join(args, ","))
	fmt.Println(args)
	panic("don't use this, use the real function")
	url := fmt.Sprintf("--url=http://"+s.kubeAddr+"/api/v1/namespaces/%s/services/squash-server:http-squash-api/proxy/api/v2", s.Namespace)
	newargs := []string{url, "--json"}
	newargs = append(newargs, args...)

	cmd := exec.Command("../../target/squash", newargs...)
	fmt.Fprintln(GinkgoWriter, "squash:", cmd.Args)

	return cmd
}
