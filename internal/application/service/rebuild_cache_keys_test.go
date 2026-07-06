package service

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

func TestStableChunkIDSameForSameContentAcrossRuns(t *testing.T) {
	got1 := stableChunkID("kid-1", types.ChunkTypeText, "hash-a", 0)
	got2 := stableChunkID("kid-1", types.ChunkTypeText, "hash-a", 0)

	require.Equal(t, got1, got2)
	require.Len(t, got1, 36)
}

func TestStableChunkIDDisambiguatesDuplicateContent(t *testing.T) {
	first := stableChunkID("kid-1", types.ChunkTypeText, "same-hash", 0)
	second := stableChunkID("kid-1", types.ChunkTypeText, "same-hash", 1)

	require.NotEqual(t, first, second)
}

func TestArtifactPromptVersionStableAndCompact(t *testing.T) {
	got1 := artifactPromptVersion("prompt", "v1")
	got2 := artifactPromptVersion("prompt", "v1")

	require.Equal(t, got1, got2)
	require.Len(t, got1, 32)
}

func TestEmbeddingArtifactKeyIncludesModelDimensionsAndCanonicalText(t *testing.T) {
	got1 := embeddingArtifactKey(7, "model-1", 1536, "cfg", "hello\r\n")
	got2 := embeddingArtifactKey(7, "model-1", 1536, "cfg", "hello\n")
	changedModel := embeddingArtifactKey(7, "model-2", 1536, "cfg", "hello\n")
	changedDims := embeddingArtifactKey(7, "model-1", 1024, "cfg", "hello\n")

	require.Equal(t, got1, got2)
	require.Equal(t, "embedding:v1", got1.ArtifactType)
	require.Equal(t, "model-1", got1.ModelID)
	require.NotEqual(t, got1.ModelID, changedModel.ModelID)
	require.NotEqual(t, got1.ConfigHash, changedDims.ConfigHash)
}

func TestSummaryArtifactKeyIncludesPromptAndLanguage(t *testing.T) {
	got := summaryArtifactKey(7, "chat-1", "zh-CN", "prompt", "cfg", "content")

	require.Equal(t, uint64(7), got.TenantID)
	require.Equal(t, "summary:v1", got.ArtifactType)
	require.Equal(t, "chat-1", got.ModelID)
	require.Equal(t, "cfg", got.ConfigHash)
	require.Equal(t, "prompt", got.PromptVersion)
}

func TestVLMImageArtifactKeyDependsOnImageBytesAndModel(t *testing.T) {
	base := vlmImageArtifactKey(7, "vlm-1", "prompt", "cfg", "default", []byte("image"))
	same := vlmImageArtifactKey(7, "vlm-1", "prompt", "cfg", "default", []byte("image"))
	changedBytes := vlmImageArtifactKey(7, "vlm-1", "prompt", "cfg", "default", []byte("other"))
	changedModel := vlmImageArtifactKey(7, "vlm-2", "prompt", "cfg", "default", []byte("image"))

	require.Equal(t, base, same)
	require.Equal(t, "vlm:image:v1", base.ArtifactType)
	require.NotEqual(t, base.CacheKey, changedBytes.CacheKey)
	require.NotEqual(t, base.ModelID, changedModel.ModelID)
}

func TestStableImageChildChunkIDDependsOnParentTypeAndContent(t *testing.T) {
	ocr := stableImageChildChunkID("parent-1", types.ChunkTypeImageOCR, "content")
	caption := stableImageChildChunkID("parent-1", types.ChunkTypeImageCaption, "content")

	require.Equal(t, ocr, stableImageChildChunkID("parent-1", types.ChunkTypeImageOCR, "content"))
	require.NotEqual(t, ocr, caption)
	require.Len(t, ocr, 36)
}

func TestQuestionArtifactKeyDependsOnPromptContextAndQuestionCount(t *testing.T) {
	base := questionArtifactKey(7, "chat-1", "zh-CN", "prompt", "cfg", "doc", "chunk", "prev", "next", 3)
	same := questionArtifactKey(7, "chat-1", "zh-CN", "prompt", "cfg", "doc", "chunk", "prev", "next", 3)
	changedContext := questionArtifactKey(7, "chat-1", "zh-CN", "prompt", "cfg", "doc", "chunk", "changed", "next", 3)
	changedCount := questionArtifactKey(7, "chat-1", "zh-CN", "prompt", "cfg", "doc", "chunk", "prev", "next", 4)

	require.Equal(t, base, same)
	require.Equal(t, "question:v1", base.ArtifactType)
	require.NotEqual(t, base.CacheKey, changedContext.CacheKey)
	require.NotEqual(t, base.ConfigHash, changedCount.ConfigHash)
}

func TestStableSummaryChunkIDDependsOnKnowledgeAndContent(t *testing.T) {
	base := stableSummaryChunkID("kid-1", "summary")
	same := stableSummaryChunkID("kid-1", "summary")
	changed := stableSummaryChunkID("kid-1", "other summary")

	require.Equal(t, base, same)
	require.NotEqual(t, base, changed)
	require.Len(t, base, 36)
}

func TestGraphChunkArtifactKeyDependsOnChunkContentAndTemplate(t *testing.T) {
	base := graphChunkArtifactKey(7, "chat-1", "prompt", "cfg", "chunk content")
	same := graphChunkArtifactKey(7, "chat-1", "prompt", "cfg", "chunk content")
	changedContent := graphChunkArtifactKey(7, "chat-1", "prompt", "cfg", "other content")
	changedPrompt := graphChunkArtifactKey(7, "chat-1", "prompt-2", "cfg", "chunk content")

	require.Equal(t, base, same)
	require.Equal(t, "graph:chunk:v1", base.ArtifactType)
	require.NotEqual(t, base.CacheKey, changedContent.CacheKey)
	require.NotEqual(t, base.PromptVersion, changedPrompt.PromptVersion)
}

func TestWikiMapArtifactKeyDependsOnDocumentLocalInputs(t *testing.T) {
	base := wikiMapArtifactKey(7, "chat-1", "zh-CN", "kb-1", "kid-1", "prompt", "cfg", "content")
	same := wikiMapArtifactKey(7, "chat-1", "zh-CN", "kb-1", "kid-1", "prompt", "cfg", "content")
	changedContent := wikiMapArtifactKey(7, "chat-1", "zh-CN", "kb-1", "kid-1", "prompt", "cfg", "other")
	changedLanguage := wikiMapArtifactKey(7, "chat-1", "en-US", "kb-1", "kid-1", "prompt", "cfg", "content")

	require.Equal(t, base, same)
	require.Equal(t, "wiki:map:v1", base.ArtifactType)
	require.NotEqual(t, base.CacheKey, changedContent.CacheKey)
	require.NotEqual(t, base.CacheKey, changedLanguage.CacheKey)
}
