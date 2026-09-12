---
name: golang-context
description: Idiomatic context propagation, deadlines, cancellation, and request-scoped values in Go.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Context

Accept `context.Context` as the first parameter for I/O or cancellable operations. Never store contexts in structs. Propagate request/job contexts through repositories and platform adapters. Add deadlines around outbound provider/API/database work when the caller has none. Do not use context values for ordinary dependencies; reserve them for request-scoped metadata. Background workflows that intentionally outlive an HTTP request must receive a workflow-owned context, not the cancelled request context.