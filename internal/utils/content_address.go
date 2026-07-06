package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func SHA256HexBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func CanonicalCacheText(s string) string {
	s = cleanInvalidUTF8ForCache(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func cleanInvalidUTF8ForCache(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		if r == 0 {
			i += size
			continue
		}
		b.WriteRune(r)
		i += size
	}
	return b.String()
}

func TextContentHash(s string) string {
	return SHA256HexBytes([]byte(CanonicalCacheText(s)))
}

func StableImageFilename(data []byte, originalNameOrExt string) string {
	ext := strings.ToLower(filepath.Ext(originalNameOrExt))
	if ext == "" {
		ext = ".png"
	}
	return SHA256HexBytes(data) + ext
}

func StableJSONHash(v any) string {
	b, _ := json.Marshal(v)
	return SHA256HexBytes(b)
}

func ArtifactCacheKey(parts ...string) string {
	return SHA256HexBytes([]byte(strings.Join(parts, "\x00")))
}
