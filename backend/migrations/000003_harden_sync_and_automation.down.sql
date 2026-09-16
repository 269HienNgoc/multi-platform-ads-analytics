DROP INDEX IF EXISTS campaign_workflow_metrics_capture_unique;
DROP INDEX IF EXISTS campaign_workflows_request_account_unique;

ALTER TABLE campaign_workflows
    DROP COLUMN IF EXISTS request_key;

DELETE FROM raw_provider_payloads
WHERE entity_kind IN ('page', 'pixel');

ALTER TABLE raw_provider_payloads
    DROP CONSTRAINT raw_provider_payloads_entity_kind_check;

ALTER TABLE raw_provider_payloads
    ADD CONSTRAINT raw_provider_payloads_entity_kind_check
    CHECK (entity_kind IN ('account', 'campaign', 'ad_group', 'ad', 'creative', 'metrics'));
