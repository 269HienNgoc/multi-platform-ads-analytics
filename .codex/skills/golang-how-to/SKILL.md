---
name: golang-how-to
description: Project Go skill orchestrator. Use on every backend Go coding, review, debugging, testing, database, API, CI, or refactoring task.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
  adapted_for: multi-platform-ads-analytics
paths:
  - "backend/**/*.go"
---

# Go skill routing

Always load this skill first for Go work, then load the relevant companion skills.

- General code: `golang-code-style`, `golang-naming`, `golang-safety`.
- Errors: `golang-error-handling`.
- Goroutines/workers/queues: `golang-concurrency` + `golang-context`.
- PostgreSQL/repositories/migrations: `golang-database` + `golang-security`.
- REST/API types: `golang-structs-interfaces` + `golang-design-patterns`; add `golang-swagger` when documenting endpoints.
- Zap/logging/metrics/tracing: `golang-observability`.
- Tests: `golang-testing`.
- Bugs/panics/races: `golang-troubleshooting` + `golang-safety`.
- CI/tooling: `golang-lint`, `golang-continuous-integration`, `golang-dependency-management`.
- Layout/refactors: `golang-project-layout`, `golang-modernize`, `golang-gopls`.
- Docs: `golang-documentation`.

Project constraints override generic advice: Go 1.24, YAML runtime config, Zap logging, PostgreSQL system of record, native VPS deployment, and no Docker.