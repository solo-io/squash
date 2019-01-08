package syncer

import (
	"context"

	"github.com/solo-io/squash/pkg/api/v1"
)

type DebugAttachmentSyncer struct{}

func (s DebugAttachmentSyncer) Sync(ctx context.Context, snap *v1.ApiSnapshot) error {
	return nil
}
