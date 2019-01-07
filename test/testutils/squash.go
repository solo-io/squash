package testutils

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils"
)

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
func (s *Squash) Attach(name, namespace, image, pod, container, processName, dbgger string) (*v1.DebugAttachment, error) {

	ctx := context.TODO() // TODO
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return nil, err
	}
	da := v1.DebugAttachment{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
		Debugger:  dbgger,
		Image:     image,
		Pod:       pod,
		Container: container,
		// DebugServerAddress: fmt.Sprintf("http://"+s.kubeAddr+"/api/v1/namespaces/%s/services/squash-server:http-squash-api/proxy/api/v2", s.Namespace),
	}
	if processName != "" {
		da.ProcessName = processName
	}
	writeOpts := clients.WriteOpts{
		Ctx:               ctx,
		OverwriteExisting: false,
	}
	return (*daClient).Write(&da, writeOpts)
}

func (s *Squash) Delete(da *v1.DebugAttachment) error {
	args := []string{"delete", da.Metadata.Name}

	cmd := s.run(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		GinkgoWriter.Write(out)
		return err
	}

	return nil
}

func (s *Squash) Wait(id string) (*v1.DebugAttachment, error) {

	cmd := s.run("wait", id, "--timeout", "90")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(GinkgoWriter, "Failed service wait:", string(out))
		return nil, err
	}

	var dbgattachment v1.DebugAttachment
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
