package utils

import (
	"context"

	"github.com/solo-io/go-utils/kubeutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	v1 "github.com/solo-io/squash/pkg/api/v1"
)

func GetDebugAttachmentClient(ctx context.Context) (v1.DebugAttachmentClient, error) {
	cfg, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, err
	}
	cache := kube.NewKubeCache(ctx)
	rcFactory := &factory.KubeResourceClientFactory{
		Crd:         v1.DebugAttachmentCrd,
		Cfg:         cfg,
		SharedCache: cache,
	}
	client, err := v1.NewDebugAttachmentClient(rcFactory)
	if err != nil {
		return nil, err
	}
	if err := client.Register(); err != nil {
		return nil, err
	}
	return client, nil
}
