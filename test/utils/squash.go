package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	models "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/options"
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
