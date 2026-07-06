package repository

import (
	"context"
	"errors"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type processArtifactCacheRepository struct {
	db *gorm.DB
}

func NewProcessArtifactCacheRepository(db *gorm.DB) interfaces.ProcessArtifactCacheRepository {
	return &processArtifactCacheRepository{db: db}
}

func (r *processArtifactCacheRepository) Put(
	ctx context.Context,
	key types.ProcessArtifactCacheKey,
	payload types.JSONMap,
) error {
	row := &types.ProcessArtifactCache{
		TenantID:      key.TenantID,
		ArtifactType:  key.ArtifactType,
		CacheKey:      key.CacheKey,
		ModelID:       key.ModelID,
		ConfigHash:    key.ConfigHash,
		PromptVersion: key.PromptVersion,
		Payload:       payload,
	}

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"},
			{Name: "artifact_type"},
			{Name: "cache_key"},
			{Name: "model_id"},
			{Name: "config_hash"},
			{Name: "prompt_version"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"payload", "updated_at"}),
	}).Create(row).Error
}

func (r *processArtifactCacheRepository) Get(
	ctx context.Context,
	key types.ProcessArtifactCacheKey,
) (types.JSONMap, bool, error) {
	var row types.ProcessArtifactCache
	err := r.db.WithContext(ctx).
		Where(
			"tenant_id = ? AND artifact_type = ? AND cache_key = ? AND model_id = ? AND config_hash = ? AND prompt_version = ?",
			key.TenantID,
			key.ArtifactType,
			key.CacheKey,
			key.ModelID,
			key.ConfigHash,
			key.PromptVersion,
		).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return row.Payload, true, nil
}
