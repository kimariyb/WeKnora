package service

import (
	"context"
	"sync"
	"testing"

	"github.com/Tencent/WeKnora/internal/models/embedding"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

type countingEmbedder struct {
	calls      int
	vectors    [][]float32
	modelID    string
	modelName  string
	dimensions int
}

func (e *countingEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	e.calls++
	if len(e.vectors) == 0 {
		return nil, nil
	}
	return e.vectors[0], nil
}

func (e *countingEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	e.calls++
	return e.vectors[:len(texts)], nil
}

func (e *countingEmbedder) BatchEmbedWithPool(ctx context.Context, model embedding.Embedder, texts []string) ([][]float32, error) {
	e.calls++
	return e.vectors[:len(texts)], nil
}

func (e *countingEmbedder) GetModelName() string {
	if e.modelName != "" {
		return e.modelName
	}
	return "counting-model"
}

func (e *countingEmbedder) GetDimensions() int {
	if e.dimensions != 0 {
		return e.dimensions
	}
	return 2
}

func (e *countingEmbedder) GetModelID() string {
	if e.modelID != "" {
		return e.modelID
	}
	return "model-1"
}

type memoryArtifactCache struct {
	mu   sync.Mutex
	rows map[types.ProcessArtifactCacheKey]types.JSONMap
}

func newMemoryArtifactCache() *memoryArtifactCache {
	return &memoryArtifactCache{rows: map[types.ProcessArtifactCacheKey]types.JSONMap{}}
}

func (m *memoryArtifactCache) Put(ctx context.Context, key types.ProcessArtifactCacheKey, payload types.JSONMap) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows[key] = payload
	return nil
}

func (m *memoryArtifactCache) Get(ctx context.Context, key types.ProcessArtifactCacheKey) (types.JSONMap, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	payload, ok := m.rows[key]
	return payload, ok, nil
}

func TestCachedEmbedderBatchUsesCacheWithoutCallingInner(t *testing.T) {
	inner := &countingEmbedder{vectors: [][]float32{{9, 9}}}
	repo := newMemoryArtifactCache()
	key := embeddingArtifactKey(7, "model-1", 2, embeddingConfigHash(inner), "hello")
	require.NoError(t, repo.Put(context.Background(), key, types.JSONMap{"vector": []float64{1, 2}}))

	cached := newCachedEmbedder(7, inner, repo)
	got, err := cached.BatchEmbedWithPool(context.Background(), cached, []string{"hello"})

	require.NoError(t, err)
	require.Equal(t, [][]float32{{1, 2}}, got)
	require.Equal(t, 0, inner.calls)
}

func TestCachedEmbedderBatchWritesMisses(t *testing.T) {
	inner := &countingEmbedder{vectors: [][]float32{{3, 4}}}
	repo := newMemoryArtifactCache()
	cached := newCachedEmbedder(7, inner, repo)

	got, err := cached.BatchEmbedWithPool(context.Background(), cached, []string{"miss"})

	require.NoError(t, err)
	require.Equal(t, [][]float32{{3, 4}}, got)
	require.Equal(t, 1, inner.calls)

	key := embeddingArtifactKey(7, "model-1", 2, embeddingConfigHash(inner), "miss")
	payload, ok, err := repo.Get(context.Background(), key)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []float64{3, 4}, payload["vector"])
}

func TestCachedEmbedderUsesCanonicalTextKey(t *testing.T) {
	inner := &countingEmbedder{vectors: [][]float32{{9, 9}}}
	repo := newMemoryArtifactCache()
	key := embeddingArtifactKey(7, "model-1", 2, embeddingConfigHash(inner), "hello\n")
	require.NoError(t, repo.Put(context.Background(), key, types.JSONMap{"vector": []any{float64(5), float64(6)}}))

	cached := newCachedEmbedder(7, inner, repo)
	got, err := cached.BatchEmbedWithPool(context.Background(), cached, []string{"hello\r\n"})

	require.NoError(t, err)
	require.Equal(t, [][]float32{{5, 6}}, got)
	require.Equal(t, 0, inner.calls)
}
