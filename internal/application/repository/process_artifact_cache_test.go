package repository_test

import (
	"context"
	"testing"

	"github.com/Tencent/WeKnora/internal/application/repository"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func processArtifactCacheTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&types.ProcessArtifactCache{}))
	return db
}

func TestProcessArtifactCacheRepositoryPutAndGet(t *testing.T) {
	db := processArtifactCacheTestDB(t)
	repo := repository.NewProcessArtifactCacheRepository(db)

	key := types.ProcessArtifactCacheKey{
		TenantID:      42,
		ArtifactType:  "embedding:v1",
		CacheKey:      "abc",
		ModelID:       "embed-1",
		ConfigHash:    "cfg",
		PromptVersion: "prompt",
	}
	require.NoError(t, repo.Put(context.Background(), key, types.JSONMap{"vector": []float64{1, 2, 3}}))

	got, ok, err := repo.Get(context.Background(), key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []any{float64(1), float64(2), float64(3)}, got["vector"])
}

func TestProcessArtifactCacheRepositoryPutUpsertsPayload(t *testing.T) {
	db := processArtifactCacheTestDB(t)
	repo := repository.NewProcessArtifactCacheRepository(db)

	key := types.ProcessArtifactCacheKey{
		TenantID:     42,
		ArtifactType: "summary:v1",
		CacheKey:     "same",
		ModelID:      "chat-1",
	}
	require.NoError(t, repo.Put(context.Background(), key, types.JSONMap{"summary": "old"}))
	require.NoError(t, repo.Put(context.Background(), key, types.JSONMap{"summary": "new"}))

	got, ok, err := repo.Get(context.Background(), key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "new", got["summary"])
}

func TestProcessArtifactCacheRepositoryGetMiss(t *testing.T) {
	db := processArtifactCacheTestDB(t)
	repo := repository.NewProcessArtifactCacheRepository(db)

	got, ok, err := repo.Get(context.Background(), types.ProcessArtifactCacheKey{
		TenantID:     42,
		ArtifactType: "embedding:v1",
		CacheKey:     "missing",
	})
	require.NoError(t, err)
	require.False(t, ok)
	require.Nil(t, got)
}
