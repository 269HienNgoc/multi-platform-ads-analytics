---
name: golang-lint
description: Go formatting, static analysis, and lint CI guidance.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go", ".github/workflows/*.yml"]
---

# Linting

At minimum require `gofmt`, `go vet`, and `go test ./...`. Prefer golangci-lint for broader checks when added to the project. Fix findings instead of broad `nolint`; any suppression must name the linter and explain why. Important categories for this codebase: err handling, shadowing, context misuse, ineffective assignments, SQL/resource leaks, complexity, and security. Keep lint versions pinned in CI so developer and CI results match.