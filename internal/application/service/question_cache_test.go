package service

import (
	"context"
	"testing"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/models/chat"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

type countingQuestionChat struct {
	calls    int
	response string
}

func (m *countingQuestionChat) Chat(
	context.Context,
	[]chat.Message,
	*chat.ChatOptions,
) (*types.ChatResponse, error) {
	m.calls++
	return &types.ChatResponse{Content: m.response}, nil
}

func (m *countingQuestionChat) ChatStream(
	context.Context,
	[]chat.Message,
	*chat.ChatOptions,
) (<-chan types.StreamResponse, error) {
	return nil, nil
}

func (m *countingQuestionChat) GetModelName() string { return "question-model" }
func (m *countingQuestionChat) GetModelID() string   { return "chat-1" }

func TestGetOrGenerateQuestionsWithCacheHitSkipsChat(t *testing.T) {
	ctx := context.WithValue(context.Background(), types.LanguageContextKey, "zh-CN")
	repo := newMemoryArtifactCache()
	model := &countingQuestionChat{response: "should not be used"}
	svc := &knowledgeService{
		config: &config.Config{Conversation: &config.ConversationConfig{
			GenerateQuestionsPrompt: "generate questions",
		}},
		processCacheRepo: repo,
	}
	key := questionArtifactKey(
		7,
		"chat-1",
		types.LanguageNameFromContext(ctx),
		artifactPromptVersion(svc.config.Conversation.GenerateQuestionsPrompt),
		svc.questionConfigHash(model),
		"doc",
		"content",
		"prev",
		"next",
		2,
	)
	require.NoError(t, repo.Put(ctx, key, types.JSONMap{"questions": []string{"cached one", "cached two"}}))

	questions, hit, err := svc.getOrGenerateQuestionsWithCache(
		ctx, 7, "chat-1", model, "content", "prev", "next", "doc", 2)

	require.NoError(t, err)
	require.True(t, hit)
	require.Equal(t, []string{"cached one", "cached two"}, questions)
	require.Equal(t, 0, model.calls)
}

func TestGetOrGenerateQuestionsWithCacheMissWritesResult(t *testing.T) {
	ctx := context.WithValue(context.Background(), types.LanguageContextKey, "zh-CN")
	repo := newMemoryArtifactCache()
	model := &countingQuestionChat{response: "1. First question?\n2. Second question?"}
	svc := &knowledgeService{
		config: &config.Config{Conversation: &config.ConversationConfig{
			GenerateQuestionsPrompt: "generate questions",
		}},
		processCacheRepo: repo,
	}

	questions, hit, err := svc.getOrGenerateQuestionsWithCache(
		ctx, 7, "chat-1", model, "content", "prev", "next", "doc", 2)

	require.NoError(t, err)
	require.False(t, hit)
	require.Equal(t, []string{"First question?", "Second question?"}, questions)
	require.Equal(t, 1, model.calls)

	key := questionArtifactKey(
		7,
		"chat-1",
		types.LanguageNameFromContext(ctx),
		artifactPromptVersion(svc.config.Conversation.GenerateQuestionsPrompt),
		svc.questionConfigHash(model),
		"doc",
		"content",
		"prev",
		"next",
		2,
	)
	payload, ok, err := repo.Get(ctx, key)
	require.NoError(t, err)
	require.True(t, ok)
	cached, ok := jsonStrings(payload["questions"])
	require.True(t, ok)
	require.Equal(t, questions, cached)
}
