package service

import (
	"fmt"

	"github.com/Tencent/WeKnora/internal/types"
	cachekey "github.com/Tencent/WeKnora/internal/utils"
	"github.com/google/uuid"
)

func textContentHash(s string) string {
	return cachekey.TextContentHash(s)
}

func stableChunkID(knowledgeID string, chunkType types.ChunkType, contentHash string, occurrence int) string {
	name := fmt.Sprintf("weknora:chunk:v1:%s:%s:%s:%d", knowledgeID, chunkType, contentHash, occurrence)
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(name)).String()
}

func stableQuestionID(chunkID string, question string, ordinal int) string {
	sum := textContentHash(fmt.Sprintf("weknora:question:v1:%s:%d:%s", chunkID, ordinal, question))
	return "q-" + sum[:16]
}

func artifactPromptVersion(parts ...string) string {
	return cachekey.ArtifactCacheKey(parts...)[:32]
}

func artifactConfigHash(v any) string {
	return cachekey.StableJSONHash(v)
}

func embeddingArtifactKey(
	tenantID uint64,
	modelID string,
	dimensions int,
	configHash string,
	text string,
) types.ProcessArtifactCacheKey {
	return types.ProcessArtifactCacheKey{
		TenantID:     tenantID,
		ArtifactType: "embedding:v1",
		CacheKey:     cachekey.ArtifactCacheKey(cachekey.CanonicalCacheText(text)),
		ModelID:      modelID,
		ConfigHash:   cachekey.ArtifactCacheKey(fmt.Sprintf("dim=%d", dimensions), configHash),
	}
}

func summaryArtifactKey(
	tenantID uint64,
	modelID string,
	language string,
	promptHash string,
	configHash string,
	content string,
) types.ProcessArtifactCacheKey {
	return types.ProcessArtifactCacheKey{
		TenantID:      tenantID,
		ArtifactType:  "summary:v1",
		CacheKey:      cachekey.ArtifactCacheKey(language, cachekey.CanonicalCacheText(content)),
		ModelID:       modelID,
		ConfigHash:    configHash,
		PromptVersion: promptHash,
	}
}

func questionArtifactKey(
	tenantID uint64,
	modelID string,
	language string,
	promptHash string,
	configHash string,
	docName string,
	content string,
	prevContent string,
	nextContent string,
	questionCount int,
) types.ProcessArtifactCacheKey {
	return types.ProcessArtifactCacheKey{
		TenantID:     tenantID,
		ArtifactType: "question:v1",
		CacheKey: cachekey.ArtifactCacheKey(
			language,
			cachekey.CanonicalCacheText(docName),
			cachekey.CanonicalCacheText(content),
			cachekey.CanonicalCacheText(prevContent),
			cachekey.CanonicalCacheText(nextContent),
		),
		ModelID:       modelID,
		ConfigHash:    cachekey.ArtifactCacheKey(fmt.Sprintf("question_count=%d", questionCount), configHash),
		PromptVersion: promptHash,
	}
}

func graphChunkArtifactKey(
	tenantID uint64,
	modelID string,
	promptHash string,
	configHash string,
	content string,
) types.ProcessArtifactCacheKey {
	return types.ProcessArtifactCacheKey{
		TenantID:      tenantID,
		ArtifactType:  "graph:chunk:v1",
		CacheKey:      cachekey.ArtifactCacheKey(cachekey.CanonicalCacheText(content)),
		ModelID:       modelID,
		ConfigHash:    configHash,
		PromptVersion: promptHash,
	}
}

func wikiMapArtifactKey(
	tenantID uint64,
	modelID string,
	language string,
	knowledgeBaseID string,
	knowledgeID string,
	promptHash string,
	configHash string,
	content string,
) types.ProcessArtifactCacheKey {
	return types.ProcessArtifactCacheKey{
		TenantID:     tenantID,
		ArtifactType: "wiki:map:v1",
		CacheKey: cachekey.ArtifactCacheKey(
			language,
			knowledgeBaseID,
			knowledgeID,
			cachekey.CanonicalCacheText(content),
		),
		ModelID:       modelID,
		ConfigHash:    configHash,
		PromptVersion: promptHash,
	}
}

func docReaderArtifactKey(
	tenantID uint64,
	fileType string,
	parserEngine string,
	configHash string,
	fileBytes []byte,
) types.ProcessArtifactCacheKey {
	return types.ProcessArtifactCacheKey{
		TenantID:     tenantID,
		ArtifactType: "docreader:v1",
		CacheKey:     cachekey.SHA256HexBytes(fileBytes),
		ModelID:      parserEngine,
		ConfigHash:   cachekey.ArtifactCacheKey(fileType, configHash),
	}
}

func vlmImageArtifactKey(
	tenantID uint64,
	modelID string,
	promptHash string,
	configHash string,
	imageSourceType string,
	imageBytes []byte,
) types.ProcessArtifactCacheKey {
	return types.ProcessArtifactCacheKey{
		TenantID:      tenantID,
		ArtifactType:  "vlm:image:v1",
		CacheKey:      cachekey.ArtifactCacheKey(imageSourceType, cachekey.SHA256HexBytes(imageBytes)),
		ModelID:       modelID,
		ConfigHash:    configHash,
		PromptVersion: promptHash,
	}
}

func stableImageChildChunkID(parentID string, chunkType types.ChunkType, content string) string {
	return stableChunkID(parentID, chunkType, textContentHash(content), 0)
}

func stableSummaryChunkID(knowledgeID string, summary string) string {
	return stableChunkID(knowledgeID, types.ChunkTypeSummary, textContentHash(summary), 0)
}
