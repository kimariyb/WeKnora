package service

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"testing"

	"github.com/Tencent/WeKnora/internal/config"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

type convertCacheFileService struct {
	files map[string][]byte
}

func newConvertCacheFileService() *convertCacheFileService {
	return &convertCacheFileService{files: map[string][]byte{}}
}

func (s *convertCacheFileService) CheckConnectivity(context.Context) error { return nil }
func (s *convertCacheFileService) SaveFile(context.Context, *multipart.FileHeader, uint64, string) (string, error) {
	return "", nil
}
func (s *convertCacheFileService) SaveBytes(context.Context, []byte, uint64, string, bool) (string, error) {
	return "", nil
}
func (s *convertCacheFileService) SaveContentAddressedBytes(context.Context, []byte, uint64, string, bool) (string, error) {
	return "", nil
}
func (s *convertCacheFileService) GetFile(_ context.Context, filePath string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.files[filePath])), nil
}
func (s *convertCacheFileService) GetFileURL(context.Context, string) (string, error) { return "", nil }
func (s *convertCacheFileService) DeleteFile(context.Context, string) error           { return nil }
func (s *convertCacheFileService) CopyFile(context.Context, string, uint64, string) (string, error) {
	return "", nil
}

type countingDocReader struct {
	calls int
}

func (r *countingDocReader) Read(context.Context, *types.ReadRequest) (*types.ReadResult, error) {
	r.calls++
	return &types.ReadResult{
		MarkdownContent: "# parsed",
		Metadata:        map[string]string{"pages": "1"},
	}, nil
}
func (r *countingDocReader) Reconnect(string) error { return nil }
func (r *countingDocReader) IsConnected() bool      { return true }
func (r *countingDocReader) ListEngines(context.Context, map[string]string) ([]types.ParserEngineInfo, error) {
	return nil, nil
}

func TestConvertFileURLDownloadedBytesReuseDocReaderCacheAcrossTempPaths(t *testing.T) {
	ctx := context.WithValue(context.Background(), types.TenantIDContextKey, uint64(42))
	fileSvc := newConvertCacheFileService()
	docReader := &countingDocReader{}
	cache := newMemoryArtifactCache()
	svc := &knowledgeService{
		config:           &config.Config{KnowledgeBase: &config.KnowledgeBaseConfig{}},
		fileSvc:          fileSvc,
		documentReader:   docReader,
		processCacheRepo: cache,
	}
	kb := &types.KnowledgeBase{
		ID:       "kb-1",
		TenantID: 42,
		ChunkingConfig: types.ChunkingConfig{
			ParserEngineRules: []types.ParserEngineRule{{FileTypes: []string{"pdf"}, Engine: "builtin"}},
		},
	}
	knowledge := &types.Knowledge{
		ID:              "kid-1",
		TenantID:        42,
		KnowledgeBaseID: kb.ID,
		Title:           "Doc",
	}
	eff := types.EffectiveProcessConfig{ChunkingConfig: kb.ChunkingConfig}

	fileSvc.files["local://42/tmp/random-a.pdf"] = []byte("same-downloaded-bytes")
	first, err := svc.convert(ctx, types.DocumentProcessPayload{
		TenantID:        42,
		KnowledgeID:     knowledge.ID,
		KnowledgeBaseID: kb.ID,
		FileURL:         "https://cdn.example.test/doc.pdf",
		FilePath:        "local://42/tmp/random-a.pdf",
		FileName:        "doc.pdf",
		FileType:        "pdf",
	}, kb, knowledge, eff, true)
	require.NoError(t, err)
	require.Equal(t, "# parsed", first.MarkdownContent)
	require.Equal(t, 1, docReader.calls)

	fileSvc.files["local://42/tmp/random-b.pdf"] = []byte("same-downloaded-bytes")
	second, err := svc.convert(ctx, types.DocumentProcessPayload{
		TenantID:        42,
		KnowledgeID:     knowledge.ID,
		KnowledgeBaseID: kb.ID,
		FileURL:         "https://cdn.example.test/doc.pdf",
		FilePath:        "local://42/tmp/random-b.pdf",
		FileName:        "doc.pdf",
		FileType:        "pdf",
	}, kb, knowledge, eff, true)
	require.NoError(t, err)
	require.Equal(t, "# parsed", second.MarkdownContent)
	require.Equal(t, 1, docReader.calls)
}
