# AGENTS.md

## Project

Multi-Platform Advertising Analytics & AI Dashboard.

## Architecture principles

- Design for multiple advertising platforms; never couple the core model to one provider.
- Canonical hierarchy: Platform -> Ad Account -> Campaign -> Ad Group/Ad Set -> Ad -> Creative.
- Prefer normalized core data plus platform-specific JSONB extensions.
- Preserve historical performance, configuration changes, raw API payloads, attribution context, and CRM/business context.
- External platform identifiers must not be used as internal database primary keys.
- PostgreSQL is the system of record.
- Backend implementation is Go.
- APIs must be suitable for REST consumers and AI/MCP analytics.
- Sync workers must be idempotent, resumable, observable, and rate-limit aware.

## Branch policy

- `main`: production branch and production deployment source.
- `dev`: integration/local testing branch.
- Create `feature/*`, `fix/*`, or other working branches from `dev`.
- Merge working branches into `dev` first.
- Only merge `dev` into `main` after localhost/integration validation succeeds.
- Avoid direct feature development on `main`.

## Documentation

Record durable architecture decisions in `docs/decisions/` and keep platform API notes under `docs/api-reference/`.
