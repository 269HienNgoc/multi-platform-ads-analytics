# ADR 0001: Campaign automation foundation

## Status

Accepted

## Context

The platform must manage multiple advertising accounts at the same time. A marketer can apply one campaign template to many accounts, but every account needs independent status, errors, spend limits, external campaign IDs and performance metrics.

The repository already establishes Go as the backend language, PostgreSQL as the source of truth and a provider-neutral canonical advertising model.

## Decision

1. Keep Go for backend services.
2. Use a provider-neutral `AdvertisingPlatform` interface; Meta-specific types stay behind the adapter.
3. Model automation as a durable workflow state machine instead of one long HTTP request.
4. Store one workflow per advertising account.
5. Store automation rules as data and evaluate them with minimum-spend/minimum-sample guardrails.
6. Keep PostgreSQL as the durable source of truth; Redis is reserved for cache/queue coordination, not canonical state.
7. Use Next.js for the operator dashboard.
8. AI produces proposals; deterministic validation/rules execute publish, pause and budget actions.
9. Backend runtime configuration and logging conventions are defined separately in ADR 0002.

## Consequences

- Bulk setup can partially succeed without rolling back unrelated accounts.
- Platform adapters can be added without changing core workflow concepts.
- Workers can be scaled independently from the API.
- We must implement idempotency, audit logging and durable PostgreSQL repositories before production execution.
