# Campaign automation

Campaign automation creates one durable workflow per advertising account. The core state machine and rule evaluator are provider-neutral; Meta, TikTok, and Google connectors implement application-owned ports.

```mermaid
stateDiagram-v2
    [*] --> SEED_PENDING
    SEED_PENDING --> SEED_CREATING
    SEED_CREATING --> SEED_RUNNING
    SEED_RUNNING --> SEED_COMPLETED
    SEED_COMPLETED --> MAIN_PENDING
    MAIN_PENDING --> MAIN_VALIDATING
    MAIN_VALIDATING --> MAIN_CREATING
    MAIN_CREATING --> MAIN_RUNNING
    MAIN_RUNNING --> PAUSED
    MAIN_RUNNING --> NEEDS_MANUAL_REVIEW
```

Metrics and state changes share a PostgreSQL transaction. When seed spend reaches its configured threshold, the service validates both required transitions and stores the final `MAIN_PENDING` state atomically.

The Meta adapter can read visible accounts and pause a named campaign. Campaign creation and budget mutation intentionally return `ErrNotImplemented` until deterministic validation, approval, idempotency, and audit-log guardrails are available.

The API currently supports workflow creation, listing, explicit transitions, and metric ingestion. Background workers, connector credential management, automation-rule CRUD, and tenant authorization remain future boundaries.
