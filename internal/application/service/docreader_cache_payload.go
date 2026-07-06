package service

import (
	"encoding/base64"

	"github.com/Tencent/WeKnora/internal/types"
)

func encodeReadResultPayload(result *types.ReadResult) types.JSONMap {
	if result == nil {
		return nil
	}
	refs := make([]any, 0, len(result.ImageRefs))
	for _, ref := range result.ImageRefs {
		refs = append(refs, types.JSONMap{
			"filename":     ref.Filename,
			"original_ref": ref.OriginalRef,
			"mime_type":    ref.MimeType,
			"storage_key":  ref.StorageKey,
			"image_data":   base64.StdEncoding.EncodeToString(ref.ImageData),
			"is_original":  ref.IsOriginal,
		})
	}
	return types.JSONMap{
		"markdown":   result.MarkdownContent,
		"metadata":   result.Metadata,
		"image_refs": refs,
		"image_dir":  result.ImageDirPath,
		"is_audio":   result.IsAudio,
		"audio_data": base64.StdEncoding.EncodeToString(result.AudioData),
	}
}

func decodeReadResultPayload(payload types.JSONMap) *types.ReadResult {
	result := &types.ReadResult{
		MarkdownContent: stringFromPayload(payload, "markdown"),
		ImageDirPath:    stringFromPayload(payload, "image_dir"),
		Metadata:        stringMapFromPayload(payload["metadata"]),
		IsAudio:         boolFromPayload(payload, "is_audio"),
	}
	if rawAudio := stringFromPayload(payload, "audio_data"); rawAudio != "" {
		if data, err := base64.StdEncoding.DecodeString(rawAudio); err == nil {
			result.AudioData = data
		}
	}
	for _, item := range anySliceFromPayload(payload["image_refs"]) {
		m, ok := item.(map[string]any)
		if !ok {
			if jm, ok := item.(types.JSONMap); ok {
				m = map[string]any(jm)
			} else {
				continue
			}
		}
		ref := types.ImageRef{
			Filename:    stringFromAny(m["filename"]),
			OriginalRef: stringFromAny(m["original_ref"]),
			MimeType:    stringFromAny(m["mime_type"]),
			StorageKey:  stringFromAny(m["storage_key"]),
			IsOriginal:  boolFromAny(m["is_original"]),
		}
		if raw := stringFromAny(m["image_data"]); raw != "" {
			if data, err := base64.StdEncoding.DecodeString(raw); err == nil {
				ref.ImageData = data
			}
		}
		result.ImageRefs = append(result.ImageRefs, ref)
	}
	return result
}

func stringFromPayload(payload types.JSONMap, key string) string {
	return stringFromAny(payload[key])
}

func stringFromAny(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func boolFromPayload(payload types.JSONMap, key string) bool {
	return boolFromAny(payload[key])
}

func boolFromAny(v any) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

func stringMapFromPayload(v any) map[string]string {
	out := map[string]string{}
	switch m := v.(type) {
	case map[string]string:
		for k, value := range m {
			out[k] = value
		}
	case map[string]any:
		for k, value := range m {
			out[k] = stringFromAny(value)
		}
	case types.JSONMap:
		for k, value := range m {
			out[k] = stringFromAny(value)
		}
	}
	return out
}

func anySliceFromPayload(v any) []any {
	switch s := v.(type) {
	case []any:
		return s
	default:
		return nil
	}
}
