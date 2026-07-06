CREATE TABLE IF NOT EXISTS process_artifact_caches (
    id BIGSERIAL PRIMARY KEY,
    tenant_id BIGINT NOT NULL,
    artifact_type VARCHAR(64) NOT NULL,
    cache_key VARCHAR(128) NOT NULL,
    model_id VARCHAR(128) NOT NULL DEFAULT '',
    config_hash VARCHAR(64) NOT NULL DEFAULT '',
    prompt_version VARCHAR(64) NOT NULL DEFAULT '',
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_process_artifact_cache_key
    ON process_artifact_caches (
        tenant_id,
        artifact_type,
        cache_key,
        model_id,
        config_hash,
        prompt_version
    );

CREATE INDEX IF NOT EXISTS idx_process_artifact_cache_tenant_type
    ON process_artifact_caches (tenant_id, artifact_type, updated_at DESC);
