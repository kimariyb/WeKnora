package utils_test

import (
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestStableImageFilenameUsesBytesAndExtension(t *testing.T) {
	got1 := utils.StableImageFilename([]byte("same-bytes"), "scan.PNG")
	got2 := utils.StableImageFilename([]byte("same-bytes"), ".png")

	require.Equal(t, got1, got2)
	require.True(t, strings.HasSuffix(got1, ".png"))
}

func TestCanonicalCacheTextNormalizesLineEndingsAndInvalidUTF8(t *testing.T) {
	got := utils.CanonicalCacheText("a\r\nb\rc" + string([]byte{0xff}))

	require.Equal(t, "a\nb\nc", got)
}
