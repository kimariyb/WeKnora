package service

import (
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

func TestReadResultPayloadRoundTripPreservesBytesAndMetadata(t *testing.T) {
	in := &types.ReadResult{
		MarkdownContent: "# doc",
		Metadata:        map[string]string{"pages": "1"},
		ImageDirPath:    "images",
		IsAudio:         true,
		AudioData:       []byte("audio"),
		ImageRefs: []types.ImageRef{{
			Filename:    "img.png",
			OriginalRef: "img.png",
			MimeType:    "image/png",
			StorageKey:  "key",
			ImageData:   []byte("image"),
			IsOriginal:  true,
		}},
	}

	got := decodeReadResultPayload(encodeReadResultPayload(in))

	require.Equal(t, in.MarkdownContent, got.MarkdownContent)
	require.Equal(t, in.Metadata, got.Metadata)
	require.Equal(t, in.ImageDirPath, got.ImageDirPath)
	require.Equal(t, in.IsAudio, got.IsAudio)
	require.Equal(t, in.AudioData, got.AudioData)
	require.Len(t, got.ImageRefs, 1)
	require.Equal(t, in.ImageRefs[0], got.ImageRefs[0])
}

func TestDocReaderArtifactKeyDependsOnFileBytesAndParserConfig(t *testing.T) {
	base := docReaderArtifactKey(7, "pdf", "builtin", "cfg", []byte("same"))
	same := docReaderArtifactKey(7, "pdf", "builtin", "cfg", []byte("same"))
	changedBytes := docReaderArtifactKey(7, "pdf", "builtin", "cfg", []byte("changed"))
	changedConfig := docReaderArtifactKey(7, "pdf", "builtin", "cfg2", []byte("same"))

	require.Equal(t, base, same)
	require.Equal(t, "docreader:v1", base.ArtifactType)
	require.NotEqual(t, base.CacheKey, changedBytes.CacheKey)
	require.NotEqual(t, base.ConfigHash, changedConfig.ConfigHash)
}
