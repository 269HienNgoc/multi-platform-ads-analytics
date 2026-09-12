---
name: golang-modernize
description: Modern Go 1.24 refactoring guidance while preserving behavior and readability.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Modern Go

The backend targets Go 1.24. Prefer current standard-library facilities and language features when they simplify code without reducing clarity. Modernization is behavior-preserving work: add/keep tests first. Do not introduce clever iterator/generic abstractions where a simple loop is clearer. Remove obsolete compatibility code only after verifying the supported Go version and deployment VPS toolchain.