ALTER TABLE raw_provider_payloads
    DROP CONSTRAINT raw_provider_payloads_entity_kind_check;

ALTER TABLE raw_provider_payloads
    ADD CONSTRAINT raw_provider_payloads_entity_kind_check
    CHECK (entity_kind IN ('account', 'campaign', 'ad_group', 'ad', 'creative', 'metrics', 'page', 'pixel'));

ALTER TABLE campaign_workflows
    ADD COLUMN request_key VARCHAR(128);

UPDATE campaign_workflows
SET request_key = id::TEXT
WHERE request_key IS NULL;

ALTER TABLE campaign_workflows
    ALTER COLUMN request_key SET NOT NULL;

CREATE UNIQUE INDEX campaign_workflows_request_account_unique
    ON campaign_workflows (organization_id, request_key, ad_account_id);

DELETE FROM campaign_workflow_metrics AS older
USING campaign_workflow_metrics AS newer
WHERE older.workflow_id = newer.workflow_id
  AND older.captured_at = newer.captured_at
  AND older.id > newer.id;

CREATE UNIQUE INDEX campaign_workflow_metrics_capture_unique
    ON campaign_workflow_metrics (workflow_id, captured_at);
