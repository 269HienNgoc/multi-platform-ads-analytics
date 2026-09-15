# ADR-0003: Campaign automation foundation

## Status

Accepted

## Context

A marketer needs to apply one campaign setup to multiple advertising accounts while preserving independent progress, errors, spend thresholds, and provider campaign identifiers for each account.

## Decision

- Keep the workflow state machine and rule evaluator in the provider-neutral domain.
- Create one durable PostgreSQL workflow per internal advertising account UUID.
- Put orchestration and persistence ports in the application layer.
- Apply metric ingestion and resulting state changes in one database transaction.
- Keep provider credentials and models inside provider adapters.
- Fail closed for campaign publishing and budget changes until deterministic guardrails and audit logging exist.
- Let AI propose configurations; only validated deterministic code may execute money-impacting actions.

## Consequences

- A failed account does not roll back another account's workflow.
- API and future workers can share the same application service and state rules.
- Meta is the first connector foundation without coupling the domain to Meta types.
- Authentication, tenant authorization, idempotent jobs, rule CRUD, and audited provider mutations are required before production exposure.
