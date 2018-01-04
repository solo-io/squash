package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/solo-io/squash/pkg/models"

	v1 "k8s.io/client-go/pkg/api/v1"
)

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

func (k *Kubectl) Run(args ...string) error {
	//	args := []string{"--namespace="+k.Namespace, "--context="k.Context}
	return k.innerunns(args)
}
func (k *Kubectl) CreateLocalRoles(yamlfile string) error {

	b, err := ioutil.ReadFile(yamlfile)
	if err != nil {
		return err
	}
	yamlcontent := string(b)
	yamlcontent = strings.Replace(yamlcontent, "ClusterRole", "Role", -1)
	buffer := bytes.NewBuffer(([]byte)(yamlcontent))
	cmd := k.innerpreparens([]string{"create", "-f", "-"})
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

func (s *Squash) Attach(image, pod, container, dbgger string) (*models.DebugAttachment, error) {

	cmd := s.run("debug-container", "--namespace="+s.Namespace, image, pod, container, dbgger)
	out, err := cmd.Output()
	if err != nil {
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

func Inittest(k *Kubectl) (func(), error) {

	if err := k.CreateNS(); err != nil {
		fmt.Println("error creating ns", err)
		return nil, err
	}
	fmt.Printf("create sutff %v \n", k)

	if err := k.Run("create", "-f", "../../target/kubernetes/squash-server.yml"); err != nil {
		return nil, err
	}
	if err := k.Run("create", "-f", "../../target/kubernetes/squash-ds.yml"); err != nil {
		return nil, err
	}
	if err := k.Run("create", "-f", "../../contrib/example/service1/service1.yml"); err != nil {
		return nil, err
	}

	return func() { k.DeleteNS() }, nil
}
