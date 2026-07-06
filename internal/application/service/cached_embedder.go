package service

import (
	"context"

	"github.com/Tencent/WeKnora/internal/logger"
	"github.com/Tencent/WeKnora/internal/models/embedding"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
)

type cachedEmbedder struct {
	tenantID uint64
	inner    embedding.Embedder
	cache    interfaces.ProcessArtifactCacheRepository
}

func wrapEmbedderWithProcessCache(
	tenantID uint64,
	inner embedding.Embedder,
	cache interfaces.ProcessArtifactCacheRepository,
) embedding.Embedder {
	if inner == nil || cache == nil {
		return inner
	}
	return newCachedEmbedder(tenantID, inner, cache)
}

func newCachedEmbedder(
	tenantID uint64,
	inner embedding.Embedder,
	cache interfaces.ProcessArtifactCacheRepository,
) embedding.Embedder {
	return &cachedEmbedder{tenantID: tenantID, inner: inner, cache: cache}
}

func (e *cachedEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	vectors, err := e.BatchEmbedWithPool(ctx, e, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, nil
	}
	return vectors[0], nil
}

func (e *cachedEmbedder) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	return e.BatchEmbedWithPool(ctx, e, texts)
}

func (e *cachedEmbedder) BatchEmbedWithPool(ctx context.Context, _ embedding.Embedder, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	missTexts := make([]string, 0, len(texts))
	missIndexes := make([]int, 0, len(texts))

	for i, text := range texts {
		key := embeddingArtifactKey(
			e.tenantID,
			e.inner.GetModelID(),
			e.inner.GetDimensions(),
			embeddingConfigHash(e.inner),
			text,
		)
		payload, ok, err := e.cache.Get(ctx, key)
		if err != nil {
			logger.Warnf(ctx, "embedding cache read failed, recomputing: %v", err)
			ok = false
		}
		if ok {
			if vector, ok := jsonVectorToFloat32(payload["vector"]); ok {
				out[i] = vector
				continue
			}
			logger.Warnf(ctx, "embedding cache payload invalid, recomputing")
		}
		missTexts = append(missTexts, text)
		missIndexes = append(missIndexes, i)
	}

	if len(missTexts) == 0 {
		return out, nil
	}

	missVectors, err := e.inner.BatchEmbedWithPool(ctx, e.inner, missTexts)
	if err != nil {
		return nil, err
	}
	for i, vector := range missVectors {
		idx := missIndexes[i]
		out[idx] = vector
		key := embeddingArtifactKey(
			e.tenantID,
			e.inner.GetModelID(),
			e.inner.GetDimensions(),
			embeddingConfigHash(e.inner),
			texts[idx],
		)
		if err := e.cache.Put(ctx, key, types.JSONMap{"vector": float32VectorToJSON(vector)}); err != nil {
			logger.Warnf(ctx, "embedding cache write failed: %v", err)
		}
	}

	return out, nil
}

func (e *cachedEmbedder) GetModelName() string { return e.inner.GetModelName() }
func (e *cachedEmbedder) GetDimensions() int   { return e.inner.GetDimensions() }
func (e *cachedEmbedder) GetModelID() string   { return e.inner.GetModelID() }

func embeddingConfigHash(model embedding.Embedder) string {
	return artifactConfigHash(types.JSONMap{
		"model_name": model.GetModelName(),
		"dimensions": model.GetDimensions(),
	})
}

func jsonVectorToFloat32(v any) ([]float32, bool) {
	switch values := v.(type) {
	case []float32:
		return append([]float32(nil), values...), true
	case []float64:
		out := make([]float32, len(values))
		for i, value := range values {
			out[i] = float32(value)
		}
		return out, true
	case []any:
		out := make([]float32, len(values))
		for i, value := range values {
			f, ok := value.(float64)
			if !ok {
				return nil, false
			}
			out[i] = float32(f)
		}
		return out, true
	default:
		return nil, false
	}
}

func float32VectorToJSON(vector []float32) []float64 {
	out := make([]float64, len(vector))
	for i, value := range vector {
		out[i] = float64(value)
	}
	return out
}
