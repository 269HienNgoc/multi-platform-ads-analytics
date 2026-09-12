# Campaign Automation Architecture

## Goal

Support many organizations, brands, pages and advertising accounts while allowing one bulk action to create independent campaign workflows per ad account.

The automation is provider-neutral. Meta is the first adapter, while TikTok/Google can implement the same core interface later.

## Runtime flow

```mermaid
flowchart TD
    U[Marketer] --> B[Bulk request]
    B --> Q[Create one workflow per ad account]
    Q --> P[Preflight validation]
    P -->|pass| S[Create seed engagement campaign]
    P -->|fail| M[Needs manual review]
    S --> I[Sync insights]
    I --> D{Seed spend >= threshold?}
    D -->|no| I
    D -->|yes| PS[Pause seed]
    PS --> V[Validate main campaign]
    V --> C[Create conversion campaign]
    C --> R[Run with initial budget]
    R --> X[Metrics + CRM events]
    X --> E[Rule engine]
    E -->|good| UP[Scale within guardrails]
    E -->|normal| R
    E -->|bad| PA[Pause]
    E -->|policy/account issue| M
    UP --> R
```

## Workflow state machine

```mermaid
stateDiagram-v2
    [*] --> ACCOUNT_CONNECTED
    ACCOUNT_CONNECTED --> ASSETS_SYNCED
    ASSETS_SYNCED --> PREFLIGHT_PASSED
    PREFLIGHT_PASSED --> SEED_PENDING
    SEED_PENDING --> SEED_CREATING
    SEED_CREATING --> SEED_RUNNING
    SEED_RUNNING --> SEED_COMPLETED
    SEED_COMPLETED --> MAIN_PENDING
    MAIN_PENDING --> MAIN_VALIDATING
    MAIN_VALIDATING --> MAIN_CREATING
    MAIN_CREATING --> MAIN_RUNNING
    MAIN_RUNNING --> PAUSED
    MAIN_RUNNING --> NEEDS_MANUAL_REVIEW
    SEED_RUNNING --> NEEDS_MANUAL_REVIEW
    NEEDS_MANUAL_REVIEW --> MAIN_PENDING
    NEEDS_MANUAL_REVIEW --> SEED_PENDING
```

Each advertising account owns its own workflow state. A failure on account B must not roll back successful work for accounts A or C.

## ERP / entity model

```mermaid
erDiagram
    ORGANIZATION ||--o{ BRAND : owns
    ORGANIZATION ||--o{ PLATFORM_CONNECTION : connects
    PLATFORM_CONNECTION ||--o{ AD_ACCOUNT : exposes
    PLATFORM_CONNECTION ||--o{ PAGE : exposes
    PAGE }o--o{ AD_ACCOUNT : usable_by
    PLATFORM_CONNECTION ||--o{ TRACKING_SOURCE : exposes
    TRACKING_SOURCE ||--o{ CONVERSION_EVENT : emits
    ORGANIZATION ||--o{ CAMPAIGN_TEMPLATE : defines
    CAMPAIGN_TEMPLATE ||--o{ CAMPAIGN_WORKFLOW : instantiates
    AD_ACCOUNT ||--o{ CAMPAIGN_WORKFLOW : runs
    PAGE ||--o{ CAMPAIGN_WORKFLOW : represents
    CAMPAIGN_WORKFLOW ||--o{ CAMPAIGN_METRIC : records
    ORGANIZATION ||--o{ AUTOMATION_RULE : controls
    ORGANIZATION ||--o{ BATCH_JOB : requests
    ORGANIZATION ||--o{ AUDIT_LOG : audits
```

## Safety and execution boundaries

AI may propose targeting, copy, creatives and campaign configuration. Deterministic code must validate and execute money-impacting operations. Budget changes, pauses and publishing should go through configured rules, limits and audit logs.

Provider policy errors, account restrictions, payment failures and ambiguous asset ownership should transition to `NEEDS_MANUAL_REVIEW`; the platform must not attempt policy-evasion behavior.

## Worker design

The next persistence/worker iteration should separate these job classes:

- `asset.sync`
- `campaign.seed.create`
- `campaign.seed.monitor`
- `campaign.main.validate`
- `campaign.main.create`
- `metrics.sync`
- `rules.evaluate`
- `campaign.pause`
- `campaign.scale`
- `creative.generate`

Jobs must be idempotent and should carry an idempotency key based on workflow + operation + version.
