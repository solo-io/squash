package kubecdl

// When you wrap kubectl, it's pronounced "kube cuddle"

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Kubecdl struct {
	Context, Namespace string

	proxyProcess *os.Process
	ProxyAddress *string

	OutputWriter io.Writer
}

func NewKubecdl(namespace string, kubectlctx string, w io.Writer) *Kubecdl {
	return &Kubecdl{
		Context:      kubectlctx,
		Namespace:    namespace,
		OutputWriter: w,
	}
}

func (k *Kubecdl) String() string {
	return fmt.Sprintf("context: %v, namespace: %v, proxyProcess: %v, ProxyAddress: %v", k.Context, k.Namespace, *k.proxyProcess, *k.ProxyAddress)
}

func (k *Kubecdl) GrantClusterAdminPermissions(clusterRoleBindingName string) error {
	args := []string{"create", "clusterrolebinding", clusterRoleBindingName, "--clusterrole=cluster-admin", fmt.Sprintf("--serviceaccount=%v:default", k.Namespace)}

	cmd := k.innerprepare(args)
	cmd.Stderr = k.OutputWriter
	cmd.Stdout = k.OutputWriter
	return cmd.Run()
}

func (k *Kubecdl) RemoveClusterAdminPermissions(clusterRoleBindingName string) error {
	args := []string{"delete", "clusterrolebinding", clusterRoleBindingName}

	cmd := k.innerprepare(args)
	cmd.Stderr = k.OutputWriter
	cmd.Stdout = k.OutputWriter
	return cmd.Run()
}

func (k *Kubecdl) CreateNS() error {
	args := []string{"create", "namespace", k.Namespace}
	return k.innerprepare(args).Run()
}

func (k *Kubecdl) DeleteDebugAttachment(name string) error {
	if k.Namespace != "" {
		args := []string{"delete", "debugattachment", name}
		return k.innerpreparens(args).Run()
	}
	return nil
}

func (k *Kubecdl) DeleteNS() error {
	if k.Namespace != "" {
		args := []string{"delete", "namespace", k.Namespace}
		return k.innerprepare(args).Run()
	}
	return nil
}

func (k *Kubecdl) CreateLocalRolesAndSleep(yamlfile string) error {
	return k.Apply(yamlfile, k.changedlocalrole, k.changenamespace, k.addsleep)
}

func (k *Kubecdl) Create(yamlfile string) error {
	return k.Apply(yamlfile, k.changenamespace)
}

func (k *Kubecdl) CreateSleep(yamlfile string) error {
	return k.Apply(yamlfile, k.changenamespace, k.addsleep)
}

func (k *Kubecdl) changedlocalrole(yamlfile string) string {
	return strings.Replace(yamlfile, "ClusterRole", "Role", -1)
}

func (k *Kubecdl) changenamespace(yamlfile string) string {
	return strings.Replace(yamlfile, "namespace: squash", "namespace: "+k.Namespace, -1)
}

func (k *Kubecdl) addsleep(yamlfile string) string {
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

func (k *Kubecdl) Apply(yamlfile string, modifier ...func(string) string) error {

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

func (k *Kubecdl) Pods() (*v1.PodList, error) {
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

func (k *Kubecdl) Exec(pod, container, cmd string, args ...string) error {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	prepareargs := []string{"exec", pod, "-c", container, "--", cmd}
	prepareargs = append(prepareargs, args...)
	return k.Prepare(prepareargs...).Run()
}

func (k *Kubecdl) ExecAsync(pod, container, cmd string, args ...string) error {
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

func (k *Kubecdl) Cp(local, remote, pod, container string) error {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	return k.Prepare("cp", local, k.Namespace+"/"+pod+":"+remote, "-c", container).Run()
}

func (k *Kubecdl) WaitPods(ctx context.Context) error {
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

func (k *Kubecdl) Proxy() error {
	cmd := exec.Command("kubectl", "proxy", "--port=0")

	portchan, err := runandreturn(cmd, proxyregex)
	if err != nil {
		return err
	}

	select {
	case port, ok := <-portchan:
		if ok {
			k.proxyProcess = cmd.Process
			k.ProxyAddress = &port[1]
			return nil
		}
		cmd.Process.Kill()
		return errors.New("can't find port")
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		return errors.New("timeout")
	}
}
func (k *Kubecdl) StopProxy() {
	if k.proxyProcess != nil {
		k.proxyProcess.Kill()
	}
}

var portregex = regexp.MustCompile(`from\s+\S+:(\d+)\s+->`)

func (k *Kubecdl) PortForward(name string) (*os.Process, string, error) {
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

func (k *Kubecdl) Logs(name string) ([]byte, error) {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	return k.Prepare("logs", name).CombinedOutput()
}

func (k *Kubecdl) Prepare(args ...string) *exec.Cmd {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	return k.innerpreparens(args)
}

func (k *Kubecdl) innerunns(args []string) error {
	return k.innerpreparens(args).Run()
}

func (k *Kubecdl) innerpreparens(args []string) *exec.Cmd {
	newargs := []string{"--namespace=" + k.Namespace}
	newargs = append(newargs, args...)
	return k.innerprepare(newargs)
}

func (k *Kubecdl) innerprepare(args []string) *exec.Cmd {
	var newargs []string
	if k.Context != "" {
		newargs = []string{"--context=" + k.Context}
	}
	newargs = append(newargs, args...)
	fmt.Fprintln(k.OutputWriter, "kubectl", strings.Join(newargs, " "))
	cmd := exec.Command("kubectl", newargs...)
	return cmd
}
