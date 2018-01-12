package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os/exec"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/squash/pkg/models"

	v1 "k8s.io/client-go/pkg/api/v1"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type Kubectl struct {
	Context, Namespace string
}

func NewKubectl(kubectlctx string) *Kubectl {
	return &Kubectl{
		Context:   kubectlctx,
		Namespace: fmt.Sprintf("test-%d", rand.Uint64()),
	}
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
	prepareargs := []string{"exec", pod, "-c", container, "--" , cmd}
	prepareargs = append(prepareargs, args...)
	return k.Prepare(prepareargs...).Run()
}

func (k *Kubectl) ExecAsync(pod, container, cmd string, args ...string) error {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	prepareargs := []string{"exec", pod, "-c", container, "--" , cmd}
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
	log.Println("kubectl", newargs)
	cmd := exec.Command("kubectl", newargs...)
	return cmd
}

func NewSquash(k *Kubectl) *Squash {
	return &Squash{Namespace: k.Namespace}
}

type Squash struct {
	Namespace string
}

func (s *Squash) Attach(image, pod, container, processName, dbgger string) (*models.DebugAttachment, error) {
	args := []string{"debug-container", "--namespace=" + s.Namespace, image, pod, container, dbgger}
	if processName != "" {
		args = append(args, "--processName="+processName)
	}

	cmd := s.run(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		GinkgoWriter.Write(out)
		return nil, err
	}

	var dbgattachment models.DebugAttachment
	err = json.Unmarshal(out, &dbgattachment)
	if err != nil {
		return nil, err
	}

	return &dbgattachment, nil
}

func (s *Squash) Wait(id string) (*models.DebugAttachment, error) {

	cmd := s.run("wait", id, "--timeout", "90")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Failed service wait:", string(out))
		return nil, err
	}

	var dbgattachment models.DebugAttachment
	err = json.Unmarshal(out, &dbgattachment)
	if err != nil {
		log.Println("Failed service wait:", string(out))
		return nil, err
	}

	return &dbgattachment, nil
}

func (s *Squash) run(args ...string) *exec.Cmd {
	url := fmt.Sprintf("--url=http://localhost:8001/api/v1/namespaces/%s/services/squash-server:http-squash-api/proxy/api/v2", s.Namespace)
	newargs := []string{url, "--json"}
	newargs = append(newargs, args...)

	cmd := exec.Command("../../target/squash", newargs...)
	log.Println("squash:", cmd.Args)

	return cmd
}
