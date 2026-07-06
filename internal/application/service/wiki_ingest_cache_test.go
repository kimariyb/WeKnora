package service

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

func TestWikiMapCachePayloadDropsRetractionsAndRestoresBaseUpdates(t *testing.T) {
	result := &docIngestResult{
		KnowledgeID: "kid-1",
		DocTitle:    "Doc",
		Summary:     "Summary",
		Pages: []types.WikiLogPageRef{
			{Slug: "summary/kid-1", Title: "Doc"},
			{Slug: "entity/acme", Title: "Acme"},
		},
		MapStats: types.JSONMap{"updates": 3},
	}
	updates := []SlugUpdate{
		{Slug: "summary/kid-1", Type: types.WikiPageTypeSummary, KnowledgeID: "kid-1"},
		{Slug: "entity/acme", Type: types.WikiPageTypeEntity, KnowledgeID: "kid-1"},
		{Slug: "entity/old", Type: "retractStale", KnowledgeID: "kid-1"},
	}

	payload := wikiMapCachePayload(result, updates)
	restoredResult, restoredUpdates, ok := wikiMapFromCachePayload(payload, nil)

	require.True(t, ok)
	require.Equal(t, "kid-1", restoredResult.KnowledgeID)
	require.Len(t, restoredResult.Pages, 2)
	require.Len(t, restoredUpdates, 2)
	require.Equal(t, "summary/kid-1", restoredUpdates[0].Slug)
	require.Equal(t, "entity/acme", restoredUpdates[1].Slug)
}

func TestAppendWikiMapRetractionsUsesCurrentOldPageSet(t *testing.T) {
	base := []SlugUpdate{
		{Slug: "summary/kid-1", Type: types.WikiPageTypeSummary, KnowledgeID: "kid-1"},
		{Slug: "entity/acme", Type: types.WikiPageTypeEntity, KnowledgeID: "kid-1"},
	}
	pages := []types.WikiLogPageRef{
		{Slug: "summary/kid-1", Title: "Doc"},
		{Slug: "entity/acme", Title: "Acme"},
	}

	updates, reparseOverlap, staleCount := appendWikiMapRetractions(
		base,
		pages,
		map[string]bool{
			"summary/kid-1": true,
			"entity/acme":   true,
			"entity/old":    true,
		},
		"prior contribution",
		"new content",
		"Doc",
		"kid-1",
		"zh-CN",
	)

	require.Equal(t, 1, reparseOverlap)
	require.Equal(t, 1, staleCount)
	require.Len(t, updates, 4)
	require.Equal(t, "retract", updates[2].Type)
	require.Equal(t, "entity/acme", updates[2].Slug)
	require.Equal(t, "prior contribution", updates[2].RetractDocContent)
	require.Equal(t, "retractStale", updates[3].Type)
	require.Equal(t, "entity/old", updates[3].Slug)
	require.Equal(t, "new content", updates[3].RetractDocContent)
}
