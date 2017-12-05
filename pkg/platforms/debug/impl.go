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
func (d *DebugPlatform) GetPid(maincontext context.Context, attachment interface{}) (int, error) {
	return 0, errors.New("This is debug mode")
}
