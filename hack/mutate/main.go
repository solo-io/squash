package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils"
)

// Purpose of script:
// This is a helper for interacting with squash CRDs

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func run() error {
	var name string
	var namespace string
	flag.StringVar(&name, "name", "some-name", "debug attachment name")
	flag.StringVar(&namespace, "namespace", "default", "debug attachment namespace")
	flag.Parse()

	mut, err := NewMutator()
	if err != nil {
		return err
	}

	if err := mut.writeResource(namespace, name); err != nil {
		return err
	}

	return nil
}

type Mutator struct {
	ctx context.Context

	daClient *v1.DebugAttachmentClient
}

func NewMutator() (Mutator, error) {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return Mutator{}, err
	}
	return Mutator{
		ctx:      ctx,
		daClient: daClient,
	}, nil
}

func (m *Mutator) writeResource(namespace, name string) error {

	da := &v1.DebugAttachment{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: namespace,
		},
		Debugger: "dlv",
	}

	wResponse, err := (*m.daClient).Write(da, clients.WriteOpts{})
	if err != nil {
		return err
	}
	fmt.Println(wResponse)
	return nil
}
