package types

import "time"

type ProcessArtifactCache struct {
	ID            int64   `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID      uint64  `json:"tenant_id" gorm:"not null;uniqueIndex:idx_process_artifact_cache_key,priority:1"`
	ArtifactType  string  `json:"artifact_type" gorm:"type:varchar(64);not null;uniqueIndex:idx_process_artifact_cache_key,priority:2"`
	CacheKey      string  `json:"cache_key" gorm:"type:varchar(128);not null;uniqueIndex:idx_process_artifact_cache_key,priority:3"`
	ModelID       string  `json:"model_id" gorm:"type:varchar(128);not null;default:'';uniqueIndex:idx_process_artifact_cache_key,priority:4"`
	ConfigHash    string  `json:"config_hash" gorm:"type:varchar(64);not null;default:'';uniqueIndex:idx_process_artifact_cache_key,priority:5"`
	PromptVersion string  `json:"prompt_version" gorm:"type:varchar(64);not null;default:'';uniqueIndex:idx_process_artifact_cache_key,priority:6"`
	Payload       JSONMap `json:"payload" gorm:"type:json;serializer:json"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ProcessArtifactCacheKey struct {
	TenantID      uint64
	ArtifactType  string
	CacheKey      string
	ModelID       string
	ConfigHash    string
	PromptVersion string
}
