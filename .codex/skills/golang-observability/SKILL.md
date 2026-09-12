---
name: golang-observability
description: Production logging, metrics, tracing, health checks, and diagnostics for this Go service.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go", "backend/config/*.yaml"]
---

# Observability

This project uses Zap, not the standard `log` package. Inject a `*zap.Logger`; prefer structured fields over formatted strings. Production logging should use JSON and `development: false`. Log lifecycle events, request method/path/status/latency, workflow/account/campaign IDs, provider operation, retry attempt, and stable error category. Never log secrets. Add metrics for provider latency/errors/rate limits, queue depth, workflow state counts, spend/insight sync lag, rule actions, and failed jobs. Keep `/health` cheap; add readiness checks when database/Redis become runtime dependencies.