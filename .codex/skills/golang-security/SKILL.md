---
name: golang-security
description: Security guidance for Go services, SQL, HTTP, secrets, tokens, and external advertising APIs.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go", "backend/config/*.yaml"]
---

# Security

Never commit real database passwords, Meta/TikTok tokens, cookies, or app secrets. Keep production `config.yaml` outside Git. Parameterize SQL; never concatenate user/provider values into queries. Apply request size/time limits and validate URLs/IDs/enum values. Redact Authorization headers and tokens from Zap logs. Encrypt stored platform credentials or use a secret manager. PostgreSQL on the same VPS should use loopback/private networking rather than public port 5432. Run `govulncheck ./...` for dependency audits. Do not build browser/cookie automation to bypass platform controls; use supported APIs and permissions.