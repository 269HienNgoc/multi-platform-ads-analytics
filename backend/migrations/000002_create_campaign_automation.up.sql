CREATE TABLE campaign_workflows (
    id UUID PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    ad_account_id UUID NOT NULL REFERENCES ad_accounts(id) ON DELETE CASCADE,
    page_external_id VARCHAR(255) NOT NULL,
    pixel_external_id VARCHAR(255) NOT NULL DEFAULT '',
    pixel_event VARCHAR(100) NOT NULL DEFAULT '',
    existing_post_id VARCHAR(255) NOT NULL DEFAULT '',
    seed_campaign_external_id VARCHAR(255) NOT NULL DEFAULT '',
    main_campaign_external_id VARCHAR(255) NOT NULL DEFAULT '',
    seed_spend_limit_usd NUMERIC(18, 6) NOT NULL DEFAULT 10 CHECK (seed_spend_limit_usd > 0),
    state VARCHAR(40) NOT NULL CHECK (state IN (
        'ACCOUNT_CONNECTED', 'ASSETS_SYNCED', 'PREFLIGHT_PASSED',
        'SEED_PENDING', 'SEED_CREATING', 'SEED_RUNNING', 'SEED_COMPLETED',
        'MAIN_PENDING', 'MAIN_VALIDATING', 'MAIN_CREATING', 'MAIN_RUNNING',
        'PAUSED', 'NEEDS_MANUAL_REVIEW', 'FAILED'
    )),
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX campaign_workflows_account_state_idx
    ON campaign_workflows (ad_account_id, state);
CREATE INDEX campaign_workflows_organization_created_idx
    ON campaign_workflows (organization_id, created_at DESC);

CREATE TABLE campaign_workflow_metrics (
    id BIGSERIAL PRIMARY KEY,
    workflow_id UUID NOT NULL REFERENCES campaign_workflows(id) ON DELETE CASCADE,
    spend_usd NUMERIC(18, 6) NOT NULL DEFAULT 0 CHECK (spend_usd >= 0),
    registrations BIGINT NOT NULL DEFAULT 0 CHECK (registrations >= 0),
    deposits BIGINT NOT NULL DEFAULT 0 CHECK (deposits >= 0),
    cost_per_registration NUMERIC(18, 6) NOT NULL DEFAULT 0 CHECK (cost_per_registration >= 0),
    cost_per_deposit NUMERIC(18, 6) NOT NULL DEFAULT 0 CHECK (cost_per_deposit >= 0),
    captured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX campaign_workflow_metrics_workflow_time_idx
    ON campaign_workflow_metrics (workflow_id, captured_at DESC);

CREATE TABLE automation_rules (
    id UUID PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    operator VARCHAR(4) NOT NULL CHECK (operator IN ('>=', '<=', '>', '<')),
    threshold NUMERIC(18, 6) NOT NULL,
    minimum_spend NUMERIC(18, 6) NOT NULL DEFAULT 0 CHECK (minimum_spend >= 0),
    minimum_samples BIGINT NOT NULL DEFAULT 0 CHECK (minimum_samples >= 0),
    action VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX automation_rules_organization_enabled_idx
    ON automation_rules (organization_id, enabled);
