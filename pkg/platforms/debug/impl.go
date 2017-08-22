package debug

import (
	"context"
	"errors"

	"github.com/solo-io/squash/pkg/platforms"
)

type DebugPlatform struct {
}

/// Watch a service and get notifications when new containers of that service are created.
func (d *DebugPlatform) WatchService(ctx context.Context, servicename string) (<-chan platforms.Container, error) {
	return nil, errors.New("This is debug mode")
}

func (d *DebugPlatform) Locate(context context.Context, containername string) (*platforms.Container, error) {
	return &platforms.Container{Image: "debug", Name: containername, Node: "debug-node"}, nil
}
func (d *DebugPlatform) GetPid(maincontext context.Context, attachmentname string) (int, error) {
	return 0, errors.New("This is debug mode")
}
