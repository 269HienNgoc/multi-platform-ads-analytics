---
name: golang-documentation
description: Go package/API documentation and durable project documentation practices.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go", "backend/**/*.md", "docs/**/*.md"]
---

# Documentation

Exported Go identifiers should have useful godoc comments when their purpose is not obvious. Document contracts, invariants, failure modes, and units rather than restating code. Keep durable architecture decisions in `docs/decisions/` and platform API notes in `docs/api-reference/`. Update README/config examples when runtime behavior changes. Never place real credentials in examples. For REST endpoints, document request/response schemas, errors, idempotency, and provider-side side effects.