---
name: golang-swagger
description: OpenAPI/Swagger documentation guidance for the Go REST API.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go", "docs/api-reference/**"]
---

# REST/OpenAPI

Keep the API contract explicit: operation purpose, path/query/body parameters, success schema, validation errors, auth requirements, and stable error codes. Document bulk operations and partial failures clearly. Campaign-changing endpoints must state whether they are drafts, approvals, or provider-side mutations. Keep generated OpenAPI artifacts reproducible and avoid hand-editing generated files. Do not expose provider access tokens or internal database fields in schemas.