---
name: golang-troubleshooting
description: Systematic debugging for Go compile errors, panics, races, leaks, API failures, and workflow bugs.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Troubleshooting

Reproduce first; reduce the failing path; add a regression test before fixing when practical. For compile failures inspect the first real compiler error, not downstream noise. For races use `go test -race`. For panics inspect stack traces and nil/ownership assumptions. For slow or memory-heavy code measure before optimizing. For campaign automation trace the workflow ID across state transition, provider request, response, retry, metric ingestion, and rule evaluation. Preserve provider request IDs/error codes in structured logs without leaking tokens.