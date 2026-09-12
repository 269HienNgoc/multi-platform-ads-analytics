---
name: golang-gopls
description: Use gopls semantic navigation and diagnostics for safe Go edits and refactors.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# gopls

Prefer semantic navigation over grep-only refactors when tools are available. Use gopls to find definitions/references, inspect package APIs, detect diagnostics, and perform safe renames. Run module download/tidy as appropriate so gopls sees the same dependency graph as CI. After refactors run `gofmt` and `go test ./...`; run race tests when concurrency changed.