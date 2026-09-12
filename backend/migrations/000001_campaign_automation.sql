BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE organizations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE brands (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name text NOT NULL,
    settings jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE platform_connections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    platform text NOT NULL,
    external_business_id text,
    status text NOT NULL DEFAULT 'ACTIVE',
    credential_ref text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ad_accounts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connection_id uuid NOT NULL REFERENCES platform_connections(id) ON DELETE CASCADE,
    platform text NOT NULL,
    external_id text NOT NULL,
    name text NOT NULL DEFAULT '',
    currency text,
    timezone text,
    status text,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (connection_id, external_id)
);

CREATE TABLE pages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connection_id uuid NOT NULL REFERENCES platform_connections(id) ON DELETE CASCADE,
    platform text NOT NULL,
    external_id text NOT NULL,
    name text NOT NULL DEFAULT '',
    status text,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (connection_id, external_id)
);

CREATE TABLE page_ad_accounts (
    page_id uuid NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    ad_account_id uuid NOT NULL REFERENCES ad_accounts(id) ON DELETE CASCADE,
    is_default boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (page_id, ad_account_id)
);

CREATE TABLE tracking_sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    connection_id uuid NOT NULL REFERENCES platform_connections(id) ON DELETE CASCADE,
    platform text NOT NULL,
    source_type text NOT NULL,
    external_id text NOT NULL,
    name text NOT NULL DEFAULT '',
    last_fired_at timestamptz,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (connection_id, source_type, external_id)
);

CREATE TABLE conversion_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_source_id uuid NOT NULL REFERENCES tracking_sources(id) ON DELETE CASCADE,
    external_key text NOT NULL,
    name text NOT NULL,
    event_type text NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tracking_source_id, external_key)
);

CREATE TABLE campaign_templates (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    brand_id uuid REFERENCES brands(id) ON DELETE SET NULL,
    name text NOT NULL,
    platform text NOT NULL,
    seed_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    main_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE campaign_workflows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    template_id uuid REFERENCES campaign_templates(id) ON DELETE SET NULL,
    ad_account_id uuid NOT NULL REFERENCES ad_accounts(id) ON DELETE CASCADE,
    page_id uuid NOT NULL REFERENCES pages(id) ON DELETE RESTRICT,
    tracking_source_id uuid REFERENCES tracking_sources(id) ON DELETE SET NULL,
    conversion_event_id uuid REFERENCES conversion_events(id) ON DELETE SET NULL,
    state text NOT NULL,
    seed_campaign_external_id text,
    main_campaign_external_id text,
    seed_spend_limit numeric(18,6) NOT NULL DEFAULT 10,
    config jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_error_code text,
    last_error_message text,
    started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX campaign_workflows_account_state_idx
    ON campaign_workflows (ad_account_id, state);

CREATE TABLE automation_rules (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name text NOT NULL,
    scope text NOT NULL,
    metric text NOT NULL,
    operator text NOT NULL,
    threshold numeric(18,6) NOT NULL,
    minimum_spend numeric(18,6) NOT NULL DEFAULT 0,
    minimum_samples integer NOT NULL DEFAULT 0,
    action text NOT NULL,
    action_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    priority integer NOT NULL DEFAULT 100,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE campaign_metrics (
    id bigserial PRIMARY KEY,
    workflow_id uuid NOT NULL REFERENCES campaign_workflows(id) ON DELETE CASCADE,
    external_campaign_id text,
    bucket_start timestamptz NOT NULL,
    spend numeric(18,6) NOT NULL DEFAULT 0,
    impressions bigint NOT NULL DEFAULT 0,
    clicks bigint NOT NULL DEFAULT 0,
    registrations bigint NOT NULL DEFAULT 0,
    deposits bigint NOT NULL DEFAULT 0,
    revenue numeric(18,6) NOT NULL DEFAULT 0,
    cost_per_registration numeric(18,6),
    cost_per_deposit numeric(18,6),
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (workflow_id, external_campaign_id, bucket_start)
);

CREATE INDEX campaign_metrics_workflow_time_idx
    ON campaign_metrics (workflow_id, bucket_start DESC);

CREATE TABLE batch_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    job_type text NOT NULL,
    status text NOT NULL DEFAULT 'PENDING',
    requested_count integer NOT NULL DEFAULT 0,
    completed_count integer NOT NULL DEFAULT 0,
    failed_count integer NOT NULL DEFAULT 0,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_logs (
    id bigserial PRIMARY KEY,
    organization_id uuid REFERENCES organizations(id) ON DELETE SET NULL,
    actor_type text NOT NULL,
    actor_id text,
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id text NOT NULL,
    before_data jsonb,
    after_data jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_logs_entity_idx ON audit_logs (entity_type, entity_id, created_at DESC);

COMMIT;
