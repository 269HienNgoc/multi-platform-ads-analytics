---
name: golang-dependency-management
description: Go module/version management, dependency review, go.mod/go.sum hygiene, and vulnerability checks.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/go.mod", "backend/go.sum", "backend/**/*.go"]
---

# Dependencies

Prefer the standard library when it is sufficient. Add dependencies for clear value, not convenience alone. Pin versions through `go.mod` and commit `go.sum`. Run `go mod tidy` after dependency changes and verify CI with `go mod download` + tests. Review module maintenance/security before adoption. Avoid hidden global initialization. Current intentional runtime dependencies include YAML parsing and Zap logging; new libraries should fit the existing architecture rather than redefine it.