package main

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils"
)

// this program exists to support the Squash development cycle
func main() {
	// namespace := options.SquashClientNamespace
	mon, err := NewMonitor()
	if err != nil {
		fmt.Println(err)
		return
	}
	mon.Run()
}

func usage() {
	use := `This is a helper utility for watching changes to squash resources.
Suggestion: run this in a side terminal while you are making changes.
Edit the print loop to show the stats you care about.
`
	fmt.Println(use)
}

type Monitor struct {
	ctx context.Context

	daClient *v1.DebugAttachmentClient
}

func NewMonitor() (Monitor, error) {
	ctx := context.Background()
	daClient, err := utils.GetDebugAttachmentClient(ctx)
	if err != nil {
		return Monitor{}, err
	}
	return Monitor{
		ctx:      ctx,
		daClient: daClient,
	}, nil
}

func (m *Monitor) Run() error {
	// setup event loop
	emitter := v1.NewApiEmitter(*m.daClient)
	syncer := m // DebugHandler implements Sync
	el := v1.NewApiEventLoop(emitter, syncer)
	// run event loop
	// watch all namespaces
	namespaces := []string{"squash", "default"}
	wOpts := clients.WatchOpts{}
	errs, err := el.Run(namespaces, wOpts)
	if err != nil {
		return err
	}
	for err := range errs {
		contextutils.LoggerFrom(m.ctx).Errorf("error in setup: %v", err)
	}
	return nil
}

func (m *Monitor) Sync(ctx context.Context, snapshot *v1.ApiSnapshot) error {
	daMap := snapshot.Debugattachments
	for _, daList := range daMap {
		for _, da := range daList {
			if err := m.syncOne(da); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Monitor) syncOne(da *v1.DebugAttachment) error {
	str := fmt.Sprintf("ns: %v, name: %v, state: %v", da.Metadata.Namespace, da.Metadata.Name, da.State)
	fmt.Println(str)
	return nil
}
