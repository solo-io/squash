package debug

import (
	"context"
	"errors"

	"github.com/solo-io/squash/pkg/platforms"
)

type DebugPlatform struct {
}

func (d *DebugPlatform) Locate(context context.Context, attachment interface{}) (interface{}, *platforms.Container, error) {
	return attachment, &platforms.Container{Image: "debug", Name: "debug", Node: "debug-node"}, nil
}
func (d *DebugPlatform) GetContainerInfo(maincontext context.Context, attachment interface{}) (*platforms.ContainerInfo, error) {
	return nil, errors.New("This is debug mode")
}
