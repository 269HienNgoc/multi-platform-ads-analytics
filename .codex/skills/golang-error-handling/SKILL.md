---
name: golang-error-handling
description: Go error propagation, wrapping, classification, and HTTP boundary handling.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Error handling

Return errors; do not log-and-return the same error at every layer. Wrap with context using `%w`, preserving `errors.Is/As`. Use sentinel/custom errors only when callers must branch on category. Convert domain errors to HTTP status codes only at the transport boundary. Include operation and stable identifiers in errors, but never secrets/access tokens. Panic only for unrecoverable programmer/startup invariants. Workers must persist/report failures in a retryable way and distinguish transient provider/rate-limit failures from permanent validation failures.