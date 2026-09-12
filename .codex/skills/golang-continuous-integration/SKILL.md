---
name: golang-continuous-integration
description: CI practices for Go builds, tests, lint/security gates, and reproducible dependency resolution.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: [".github/workflows/*.yml", "backend/**"]
---

# Go CI

CI must start from committed `go.mod`/`go.sum`; do not mutate modules during normal test jobs. Run `go mod download` then `go test ./...`. Add `go vet`, golangci-lint, race tests, and `govulncheck` as the project matures. Keep Go version aligned with `go.mod` (currently 1.24). Cache dependencies but never rely on cache correctness. Feature branches target `dev`; only validated `dev` moves to `main`. No Docker build steps should be introduced unless the project deployment strategy explicitly changes.