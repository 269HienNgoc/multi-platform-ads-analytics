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

## Go development

Before any Go coding, review, debugging, troubleshooting, or setup task, load the `samber/cc-skills-golang@golang-how-to` skill first — it routes to whichever other Go skills the task needs.

Project requirements override community skill defaults when they conflict:

- Use Clean Architecture with manual constructor injection. Keep the composition root in `backend/internal/app` and executable entry points minimal.
- Load YAML configuration through Viper in `backend/internal/config`. Environment variables may override YAML, but application packages must not read environment variables directly.
- Use Zap for structured logging. Do not use the standard `log` package or `log/slog` in backend application code.
- Use GORM for PostgreSQL persistence. Do not use raw `database/sql`, `sqlx`, or direct driver queries in application code. Accessing GORM's underlying pool is allowed only inside the PostgreSQL adapter for pool configuration, ping, and shutdown.
- Do not add Docker or Docker Compose assets unless the user explicitly changes this requirement.
- Keep domain models free of Gin, GORM, Viper, Zap, and provider SDK dependencies.
