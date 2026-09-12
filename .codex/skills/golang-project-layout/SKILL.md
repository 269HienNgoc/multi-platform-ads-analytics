---
name: golang-project-layout
description: Project/package structure guidance for the Go backend.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**"]
---

# Project layout

Keep executables under `cmd/` and application packages under `internal/`. Organize primarily by domain/capability, not generic technical buckets. Current useful boundaries include config, logging, HTTP transport, workflow, automation/rules, platform adapters, persistence/repositories, and workers. Avoid a giant `utils` package. Keep YAML runtime configuration under `backend/config`; production secrets live only in the uncommitted `config.yaml` on the VPS. The project deploys as native Go/Node processes and intentionally does not use Docker.