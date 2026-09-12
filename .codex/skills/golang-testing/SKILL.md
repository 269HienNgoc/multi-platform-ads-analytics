---
name: golang-testing
description: Production Go tests for domain logic, HTTP handlers, workers, adapters, repositories, and concurrency.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Testing

Use table-driven tests for deterministic business rules and state transitions. Test exported behavior, not implementation details. Use `httptest` for handlers. Keep platform APIs behind interfaces so tests use fakes/stubs without network calls. Cover idempotency, retry classification, partial bulk failure, rule minimum-data guards, invalid state transitions, cancellation, and duplicate provider events. Use `go test ./...`; use `go test -race ./...` for concurrency changes. Add integration tests for PostgreSQL repositories once persistence is wired. Prefer deterministic clocks/IDs via injected interfaces when time or randomness affects behavior.