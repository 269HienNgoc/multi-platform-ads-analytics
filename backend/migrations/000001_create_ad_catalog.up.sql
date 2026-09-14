CREATE TABLE platforms (
    code VARCHAR(32) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO platforms (code, name) VALUES
    ('meta', 'Meta Ads'),
    ('tiktok', 'TikTok Ads'),
    ('google', 'Google Ads')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE ad_accounts (
    id UUID PRIMARY KEY,
    platform_code VARCHAR(32) NOT NULL REFERENCES platforms(code) ON UPDATE CASCADE ON DELETE RESTRICT,
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    currency CHAR(3) NOT NULL,
    timezone VARCHAR(100) NOT NULL,
    status VARCHAR(32) NOT NULL CHECK (status IN ('active', 'paused', 'archived')),
    provider_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ad_accounts_platform_external_unique UNIQUE (platform_code, external_id)
);

CREATE INDEX ad_accounts_platform_status_idx ON ad_accounts (platform_code, status);

CREATE TABLE campaigns (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES ad_accounts(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    objective VARCHAR(100) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL CHECK (status IN ('active', 'paused', 'archived')),
    provider_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT campaigns_account_external_unique UNIQUE (account_id, external_id)
);

CREATE INDEX campaigns_account_status_idx ON campaigns (account_id, status);

CREATE TABLE ad_groups (
    id UUID PRIMARY KEY,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL CHECK (status IN ('active', 'paused', 'archived')),
    provider_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ad_groups_campaign_external_unique UNIQUE (campaign_id, external_id)
);

CREATE INDEX ad_groups_campaign_status_idx ON ad_groups (campaign_id, status);

CREATE TABLE ads (
    id UUID PRIMARY KEY,
    ad_group_id UUID NOT NULL REFERENCES ad_groups(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(32) NOT NULL CHECK (status IN ('active', 'paused', 'archived')),
    provider_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT ads_group_external_unique UNIQUE (ad_group_id, external_id)
);

CREATE INDEX ads_group_status_idx ON ads (ad_group_id, status);

CREATE TABLE creatives (
    id UUID PRIMARY KEY,
    ad_id UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    format VARCHAR(50) NOT NULL,
    asset_url TEXT NOT NULL DEFAULT '',
    provider_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT creatives_ad_external_unique UNIQUE (ad_id, external_id)
);

CREATE INDEX creatives_ad_idx ON creatives (ad_id);

CREATE TABLE performance_metrics_daily (
    id UUID PRIMARY KEY,
    platform_code VARCHAR(32) NOT NULL REFERENCES platforms(code) ON UPDATE CASCADE ON DELETE RESTRICT,
    account_id UUID NOT NULL REFERENCES ad_accounts(id) ON DELETE CASCADE,
    entity_kind VARCHAR(32) NOT NULL CHECK (entity_kind IN ('account', 'campaign', 'ad_group', 'ad', 'creative')),
    entity_id UUID NOT NULL,
    metric_date DATE NOT NULL,
    impressions BIGINT NOT NULL DEFAULT 0 CHECK (impressions >= 0),
    reach BIGINT NOT NULL DEFAULT 0 CHECK (reach >= 0),
    clicks BIGINT NOT NULL DEFAULT 0 CHECK (clicks >= 0),
    conversions NUMERIC(20, 6) NOT NULL DEFAULT 0 CHECK (conversions >= 0),
    spend NUMERIC(20, 6) NOT NULL DEFAULT 0 CHECK (spend >= 0),
    revenue NUMERIC(20, 6) NOT NULL DEFAULT 0 CHECK (revenue >= 0),
    provider_metrics JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT performance_metrics_entity_date_unique UNIQUE (entity_kind, entity_id, metric_date)
);

CREATE INDEX performance_metrics_account_date_idx ON performance_metrics_daily (account_id, metric_date DESC);
CREATE INDEX performance_metrics_entity_date_idx ON performance_metrics_daily (entity_kind, entity_id, metric_date DESC);

CREATE TABLE sync_runs (
    id UUID PRIMARY KEY,
    platform_code VARCHAR(32) NOT NULL REFERENCES platforms(code) ON UPDATE CASCADE ON DELETE RESTRICT,
    account_id UUID REFERENCES ad_accounts(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
    cursor_value TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    records_processed BIGINT NOT NULL DEFAULT 0 CHECK (records_processed >= 0),
    error_summary TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX sync_runs_account_created_idx ON sync_runs (account_id, created_at DESC);
CREATE INDEX sync_runs_status_created_idx ON sync_runs (status, created_at);

CREATE TABLE raw_provider_payloads (
    id UUID PRIMARY KEY,
    sync_run_id UUID NOT NULL REFERENCES sync_runs(id) ON DELETE CASCADE,
    entity_kind VARCHAR(32) NOT NULL CHECK (entity_kind IN ('account', 'campaign', 'ad_group', 'ad', 'creative', 'metrics')),
    external_id VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT raw_provider_payloads_run_hash_unique UNIQUE (sync_run_id, payload_hash)
);

CREATE INDEX raw_provider_payloads_external_idx ON raw_provider_payloads (entity_kind, external_id, received_at DESC);
