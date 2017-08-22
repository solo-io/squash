package e2e_test

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os/exec"

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

func (s *Squash) Attach(image, pod, container, dbgger string) (*models.DebugConfig, error) {

	cmd := s.run("debug-container", image, pod, container, dbgger)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var dbgconfig models.DebugConfig
	err = json.Unmarshal(out, &dbgconfig)
	if err != nil {
		return nil, err
	}

	return &dbgconfig, nil
}

func (s *Squash) Service(image, service, dbgger string, bp ...string) (*models.DebugConfig, error) {
	cmdline := []string{"debug-service", service, image, dbgger}
	for _, b := range bp {
		cmdline = append(cmdline, "--breakpoint="+b)
	}

	cmd := s.run(cmdline...)
	out, err := cmd.Output()
	if err != nil {
		log.Println("Failed service attach:", string(out))
		return nil, err
	}

	var dbgconfig models.DebugConfig
	err = json.Unmarshal(out, &dbgconfig)
	if err != nil {
		log.Println("Failed service attach:", string(out))
		return nil, err
	}

	return &dbgconfig, nil
}

func (s *Squash) Wait(id string) (*models.DebugSession, error) {

	cmd := s.run("wait", id, "--timeout", "90")

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Failed service wait:", string(out))
		return nil, err
	}

	var dbgsession models.DebugSession
	err = json.Unmarshal(out, &dbgsession)
	if err != nil {
		log.Println("Failed service wait:", string(out))
		return nil, err
	}

	return &dbgsession, nil
}

func (s *Squash) run(args ...string) *exec.Cmd {
	url := fmt.Sprintf("--url=http://localhost:8001/api/v1/namespaces/%s/services/squash-server-service/proxy/api/v1", s.Namespace)
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
