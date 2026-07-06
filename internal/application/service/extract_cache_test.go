package service

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

func TestGraphDataCachePayloadStripsChunkRefsAndRestoresGraph(t *testing.T) {
	graph := &types.GraphData{
		Node: []*types.GraphNode{
			{Name: "alpha", Chunks: []string{"old-chunk"}, Attributes: []string{"attr"}},
		},
		Relation: []*types.GraphRelation{
			{Node1: "alpha", Node2: "beta", Type: "relates_to"},
		},
	}

	payload := graphDataCachePayload(graph)
	restored, ok := graphDataFromCachePayload(payload)

	require.True(t, ok)
	require.Len(t, restored.Node, 1)
	require.Equal(t, "alpha", restored.Node[0].Name)
	require.Empty(t, restored.Node[0].Chunks)
	require.Len(t, restored.Relation, 1)
	require.Equal(t, "relates_to", restored.Relation[0].Type)
}
