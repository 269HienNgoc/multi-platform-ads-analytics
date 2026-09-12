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

## Required Go skills

For any change under `backend/`, load `.codex/skills/golang-how-to/SKILL.md` first and then the relevant Go companion skills.

Common routing:

- Go implementation/review: `golang-code-style`, `golang-naming`, `golang-safety`.
- Errors: `golang-error-handling`.
- Workers/goroutines: `golang-concurrency` + `golang-context`.
- PostgreSQL/repositories/migrations: `golang-database` + `golang-security`.
- Types/adapters/architecture: `golang-structs-interfaces` + `golang-design-patterns`.
- Zap/logging/metrics: `golang-observability`.
- Tests/debugging: `golang-testing`, `golang-troubleshooting`.
- Tooling/CI: `golang-lint`, `golang-dependency-management`, `golang-continuous-integration`.
- Layout/refactoring: `golang-project-layout`, `golang-modernize`, `golang-gopls`.
- REST documentation: `golang-swagger`, `golang-documentation`.

Project-specific conventions override generic skill advice: Go 1.24, YAML config, Zap logging, PostgreSQL system of record, native VPS deployment, and no Docker.

## Branch policy

- `main`: production branch and production deployment source.
- `dev`: integration/local testing branch.
- Create `feature/*`, `fix/*`, or other working branches from `dev`.
- Merge working branches into `dev` first.
- Only merge `dev` into `main` after localhost/integration validation succeeds.
- Avoid direct feature development on `main`.

## Documentation

Record durable architecture decisions in `docs/decisions/` and keep platform API notes under `docs/api-reference/`.
