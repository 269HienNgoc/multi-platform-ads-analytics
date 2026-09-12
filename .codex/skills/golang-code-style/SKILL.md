---
name: golang-code-style
description: Idiomatic Go code style and clarity rules for this backend.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Go code style

Prefer clear, boring Go. Use `gofmt`; use `:=` for non-zero local values and `var` for intentional zero values. Use named fields in composite literals. Handle errors/edge cases early and keep the happy path shallow. Avoid unnecessary `else` after return/break/continue. Keep functions focused and prefer an options/config struct over long parameter lists. Pass `context.Context` first. Avoid premature abstractions and minimize exported surface area. Initialize maps before writes; for JSON APIs prefer initialized slices when callers expect `[]` rather than `null`. Keep provider-specific behavior behind adapters instead of leaking it into domain packages.