package interfaces

import (
	"context"

	"github.com/Tencent/WeKnora/internal/types"
)

type ProcessArtifactCacheRepository interface {
	Put(ctx context.Context, key types.ProcessArtifactCacheKey, payload types.JSONMap) error
	Get(ctx context.Context, key types.ProcessArtifactCacheKey) (types.JSONMap, bool, error)
}
