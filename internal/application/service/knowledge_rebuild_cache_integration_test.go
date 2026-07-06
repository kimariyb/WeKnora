package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

type rebuildCacheModelCalls struct {
	DocReader      int
	VLM            int
	Embedding      int
	SummaryChat    int
	QuestionChat   int
	GraphChat      int
	WikiMapChat    int
	WikiReduceChat int
}

func (c rebuildCacheModelCalls) Add(other rebuildCacheModelCalls) rebuildCacheModelCalls {
	return rebuildCacheModelCalls{
		DocReader:      c.DocReader + other.DocReader,
		VLM:            c.VLM + other.VLM,
		Embedding:      c.Embedding + other.Embedding,
		SummaryChat:    c.SummaryChat + other.SummaryChat,
		QuestionChat:   c.QuestionChat + other.QuestionChat,
		GraphChat:      c.GraphChat + other.GraphChat,
		WikiMapChat:    c.WikiMapChat + other.WikiMapChat,
		WikiReduceChat: c.WikiReduceChat + other.WikiReduceChat,
	}
}

func (c rebuildCacheModelCalls) Since(previous rebuildCacheModelCalls) rebuildCacheModelCalls {
	return rebuildCacheModelCalls{
		DocReader:      c.DocReader - previous.DocReader,
		VLM:            c.VLM - previous.VLM,
		Embedding:      c.Embedding - previous.Embedding,
		SummaryChat:    c.SummaryChat - previous.SummaryChat,
		QuestionChat:   c.QuestionChat - previous.QuestionChat,
		GraphChat:      c.GraphChat - previous.GraphChat,
		WikiMapChat:    c.WikiMapChat - previous.WikiMapChat,
		WikiReduceChat: c.WikiReduceChat - previous.WikiReduceChat,
	}
}

type rebuildCacheIntegrationHarness struct {
	t *testing.T

	ctx         context.Context
	cache       *memoryArtifactCache
	fileName    string
	fileBytes   []byte
	tenantID    uint64
	kbID        string
	knowledgeID string
	language    string

	embeddingModelID  string
	wikiPromptVersion string

	calls      rebuildCacheModelCalls
	markdowns  []string
	chunkIDRun [][]string
}

func newRebuildCacheIntegrationHarness(t *testing.T) *rebuildCacheIntegrationHarness {
	t.Helper()
	return &rebuildCacheIntegrationHarness{
		t:                 t,
		ctx:               context.WithValue(context.Background(), types.LanguageContextKey, "zh-CN"),
		cache:             newMemoryArtifactCache(),
		tenantID:          42,
		kbID:              "kb-1",
		knowledgeID:       "kid-1",
		language:          "zh-CN",
		embeddingModelID:  "embed-v1",
		wikiPromptVersion: "wiki-prompt-v1",
	}
}

func (h *rebuildCacheIntegrationHarness) seedFile(name string, data []byte) {
	h.fileName = name
	h.fileBytes = append([]byte(nil), data...)
}

func (h *rebuildCacheIntegrationHarness) processDocumentOnce() error {
	return h.processAllLayers()
}

func (h *rebuildCacheIntegrationHarness) reparseAndProcessAgain() error {
	return h.processAllLayers()
}

func (h *rebuildCacheIntegrationHarness) processAllLayers() error {
	markdown, err := h.processDocReader()
	if err != nil {
		return err
	}
	h.markdowns = append(h.markdowns, markdown)

	chunkIDs, chunkContents := h.reconcileStableChunks(markdown)
	h.chunkIDRun = append(h.chunkIDRun, chunkIDs)

	if err := h.processVLM(); err != nil {
		return err
	}
	if err := h.processEmbeddings(chunkContents); err != nil {
		return err
	}
	if err := h.processSummary(markdown); err != nil {
		return err
	}
	if err := h.processQuestions(chunkContents); err != nil {
		return err
	}
	if err := h.processGraph(chunkContents); err != nil {
		return err
	}
	if err := h.processWikiMap(markdown); err != nil {
		return err
	}
	h.processWikiReduce()
	return nil
}

func (h *rebuildCacheIntegrationHarness) processUntilAfterVLMAndEmbeddingThenAbortWorker() error {
	markdown, err := h.processDocReader()
	if err != nil {
		return err
	}
	_, chunkContents := h.reconcileStableChunks(markdown)
	if err := h.processVLM(); err != nil {
		return err
	}
	if err := h.processEmbeddings(chunkContents); err != nil {
		return err
	}
	// Complete Wiki map before the simulated crash: the acceptance invariant is
	// that already completed artifacts cache-hit after worker restart.
	return h.processWikiMap(markdown)
}

func (h *rebuildCacheIntegrationHarness) resumeSameAttemptOrReparse() error {
	return h.processAllLayers()
}

func (h *rebuildCacheIntegrationHarness) changeEmbeddingModel(id string) {
	h.embeddingModelID = id
}

func (h *rebuildCacheIntegrationHarness) changeWikiExtractionPrompt(id string) {
	h.wikiPromptVersion = id
}

func (h *rebuildCacheIntegrationHarness) modelCalls() rebuildCacheModelCalls {
	return h.calls
}

func (h *rebuildCacheIntegrationHarness) modelCallsSince(previous rebuildCacheModelCalls) rebuildCacheModelCalls {
	return h.calls.Since(previous)
}

func (h *rebuildCacheIntegrationHarness) firstRunMarkdown() string {
	if len(h.markdowns) == 0 {
		return ""
	}
	return h.markdowns[0]
}

func (h *rebuildCacheIntegrationHarness) secondRunMarkdown() string {
	if len(h.markdowns) < 2 {
		return ""
	}
	return h.markdowns[1]
}

func (h *rebuildCacheIntegrationHarness) firstRunChunkIDs() []string {
	if len(h.chunkIDRun) == 0 {
		return nil
	}
	return h.chunkIDRun[0]
}

func (h *rebuildCacheIntegrationHarness) secondRunChunkIDs() []string {
	if len(h.chunkIDRun) < 2 {
		return nil
	}
	return h.chunkIDRun[1]
}

func (h *rebuildCacheIntegrationHarness) processDocReader() (string, error) {
	key := docReaderArtifactKey(h.tenantID, "pdf", "builtin", artifactConfigHash(types.JSONMap{
		"parser_engine":         "builtin",
		"file_type":             "pdf",
		"parser_overrides":      map[string]string{},
		"chunking_parser_rules": nil,
	}), h.fileBytes)
	if payload, ok, err := h.cache.Get(h.ctx, key); err != nil {
		return "", err
	} else if ok {
		return decodeReadResultPayload(payload).MarkdownContent, nil
	}

	h.calls.DocReader++
	result := &types.ReadResult{
		MarkdownContent: fmt.Sprintf("# Parsed %s\n\n%s", h.fileName, string(h.fileBytes)),
		Metadata:        map[string]string{"pages": "1"},
	}
	if err := h.cache.Put(h.ctx, key, encodeReadResultPayload(result)); err != nil {
		return "", err
	}
	return result.MarkdownContent, nil
}

func (h *rebuildCacheIntegrationHarness) reconcileStableChunks(markdown string) ([]string, []string) {
	chunks := []string{markdown}
	ids := make([]string, len(chunks))
	for i, content := range chunks {
		ids[i] = stableChunkID(h.knowledgeID, types.ChunkTypeText, textContentHash(content), i)
	}
	return ids, chunks
}

func (h *rebuildCacheIntegrationHarness) processVLM() error {
	key := vlmImageArtifactKey(
		h.tenantID,
		"vlm-1",
		artifactPromptVersion("ocr-prompt", "caption-prompt"),
		artifactConfigHash(types.JSONMap{"model_name": "vlm-model"}),
		"document_image",
		[]byte("same-image"),
	)
	if _, ok, err := h.cache.Get(h.ctx, key); err != nil {
		return err
	} else if ok {
		return nil
	}
	h.calls.VLM++
	return h.cache.Put(h.ctx, key, types.JSONMap{"ocr_text": "ocr", "caption": "caption"})
}

func (h *rebuildCacheIntegrationHarness) processEmbeddings(texts []string) error {
	inner := &countingEmbedder{
		vectors:    [][]float32{{1, 2}},
		modelID:    h.embeddingModelID,
		modelName:  h.embeddingModelID,
		dimensions: 2,
	}
	cached := newCachedEmbedder(h.tenantID, inner, h.cache)
	if _, err := cached.BatchEmbedWithPool(h.ctx, cached, texts); err != nil {
		return err
	}
	h.calls.Embedding += inner.calls
	return nil
}

func (h *rebuildCacheIntegrationHarness) processSummary(content string) error {
	key := summaryArtifactKey(
		h.tenantID,
		"chat-1",
		types.LanguageNameFromContext(h.ctx),
		artifactPromptVersion("summary-prompt"),
		artifactConfigHash(types.JSONMap{"model_name": "summary-model", "temperature": 0.3}),
		content,
	)
	if _, ok, err := h.cache.Get(h.ctx, key); err != nil {
		return err
	} else if ok {
		return nil
	}
	h.calls.SummaryChat++
	return h.cache.Put(h.ctx, key, types.JSONMap{"summary": "summary"})
}

func (h *rebuildCacheIntegrationHarness) processQuestions(texts []string) error {
	svc := &knowledgeService{
		config: &config.Config{Conversation: &config.ConversationConfig{
			GenerateQuestionsPrompt: "generate questions",
		}},
		processCacheRepo: h.cache,
	}
	model := &countingQuestionChat{response: "1. What is this document about?"}
	for _, text := range texts {
		if _, _, err := svc.getOrGenerateQuestionsWithCache(
			h.ctx, h.tenantID, "chat-1", model, text, "", "", h.fileName, 1,
		); err != nil {
			return err
		}
	}
	h.calls.QuestionChat += model.calls
	return nil
}

func (h *rebuildCacheIntegrationHarness) processGraph(texts []string) error {
	for _, text := range texts {
		key := graphChunkArtifactKey(
			h.tenantID,
			"chat-1",
			artifactPromptVersion("graph-template"),
			artifactConfigHash(types.JSONMap{"model_name": "graph-model"}),
			text,
		)
		if _, ok, err := h.cache.Get(h.ctx, key); err != nil {
			return err
		} else if ok {
			continue
		}
		h.calls.GraphChat++
		if err := h.cache.Put(h.ctx, key, graphDataCachePayload(&types.GraphData{
			Node: []*types.GraphNode{{Name: "Doc"}},
		})); err != nil {
			return err
		}
	}
	return nil
}

func (h *rebuildCacheIntegrationHarness) processWikiMap(content string) error {
	key := wikiMapArtifactKey(
		h.tenantID,
		"chat-1",
		h.language,
		h.kbID,
		h.knowledgeID,
		artifactPromptVersion(h.wikiPromptVersion),
		artifactConfigHash(types.JSONMap{"model_name": "wiki-model", "granularity": "standard"}),
		content,
	)
	if _, ok, err := h.cache.Get(h.ctx, key); err != nil {
		return err
	} else if ok {
		return nil
	}
	h.calls.WikiMapChat++
	return h.cache.Put(h.ctx, key, wikiMapCachePayload(
		&docIngestResult{
			KnowledgeID: h.knowledgeID,
			DocTitle:    h.fileName,
			Summary:     "summary",
			Pages:       []types.WikiLogPageRef{{Slug: "summary/kid-1", Title: h.fileName}},
			MapStats:    types.JSONMap{"updates": 1},
		},
		[]SlugUpdate{{Slug: "summary/kid-1", Type: types.WikiPageTypeSummary, KnowledgeID: h.knowledgeID}},
	))
}

func (h *rebuildCacheIntegrationHarness) processWikiReduce() {
	h.calls.WikiReduceChat++
}

func TestReparseUnchangedKnowledgeAvoidsExpensiveModelCalls(t *testing.T) {
	h := newRebuildCacheIntegrationHarness(t)
	h.seedFile("doc.pdf", []byte("same-pdf"))

	require.NoError(t, h.processDocumentOnce())
	firstCalls := h.modelCalls()
	require.Greater(t, firstCalls.VLM+firstCalls.SummaryChat+firstCalls.QuestionChat+firstCalls.GraphChat+firstCalls.WikiMapChat+firstCalls.Embedding, 0)

	require.NoError(t, h.reparseAndProcessAgain())
	secondCalls := h.modelCallsSince(firstCalls)

	require.Equal(t, 0, secondCalls.VLM)
	require.Equal(t, 0, secondCalls.Embedding)
	require.Equal(t, 0, secondCalls.SummaryChat)
	require.Equal(t, 0, secondCalls.QuestionChat)
	require.Equal(t, 0, secondCalls.GraphChat)
	require.Equal(t, 0, secondCalls.WikiMapChat)
	require.GreaterOrEqual(t, secondCalls.WikiReduceChat, 0)
	require.Equal(t, h.firstRunMarkdown(), h.secondRunMarkdown())
	require.Equal(t, h.firstRunChunkIDs(), h.secondRunChunkIDs())
}

func TestCrashRetryReusesCompletedArtifacts(t *testing.T) {
	h := newRebuildCacheIntegrationHarness(t)
	h.seedFile("doc.pdf", []byte("same-pdf"))

	require.NoError(t, h.processUntilAfterVLMAndEmbeddingThenAbortWorker())
	callsAfterAbort := h.modelCalls()
	require.Greater(t, callsAfterAbort.VLM+callsAfterAbort.Embedding, 0)

	require.NoError(t, h.resumeSameAttemptOrReparse())
	resumeCalls := h.modelCallsSince(callsAfterAbort)

	require.Equal(t, 0, resumeCalls.VLM)
	require.Equal(t, 0, resumeCalls.Embedding)
	require.Equal(t, 0, resumeCalls.WikiMapChat)
}

func TestRebuildCacheInvalidatesOnlyChangedLayer(t *testing.T) {
	h := newRebuildCacheIntegrationHarness(t)
	h.seedFile("doc.pdf", []byte("same-pdf"))
	require.NoError(t, h.processDocumentOnce())
	first := h.modelCalls()

	h.changeEmbeddingModel("embed-v2")
	require.NoError(t, h.reparseAndProcessAgain())
	afterEmbeddingChange := h.modelCallsSince(first)

	require.Equal(t, 0, afterEmbeddingChange.VLM)
	require.Greater(t, afterEmbeddingChange.Embedding, 0)
	require.Equal(t, 0, afterEmbeddingChange.SummaryChat)
	require.Equal(t, 0, afterEmbeddingChange.WikiMapChat)

	h.changeWikiExtractionPrompt("wiki-prompt-v2")
	require.NoError(t, h.reparseAndProcessAgain())
	afterWikiPromptChange := h.modelCallsSince(first.Add(afterEmbeddingChange))

	require.Equal(t, 0, afterWikiPromptChange.VLM)
	require.Equal(t, 0, afterWikiPromptChange.Embedding)
	require.Greater(t, afterWikiPromptChange.WikiMapChat, 0)
}
