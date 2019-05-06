package actions

import (
	"context"

	v1 "github.com/solo-io/squash/pkg/api/v1"
	"github.com/solo-io/squash/pkg/utils"
)

type UserController struct {
	ctx      context.Context
	daClient v1.DebugAttachmentClient
}

func NewUserController(kubeconfigPath string) (UserController, error) {
	ctx := context.Background()
	daClient, err := utils.GetBasicDebugAttachmentClient(ctx, kubeconfigPath)
	if err != nil {
		return UserController{}, err
	}
	return UserController{
		ctx:      ctx,
		daClient: daClient,
	}, nil
}
