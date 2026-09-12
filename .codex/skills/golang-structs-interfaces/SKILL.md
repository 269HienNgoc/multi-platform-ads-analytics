---
name: golang-structs-interfaces
description: Go struct/interface design, receivers, tags, composition, and provider boundaries.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Structs and interfaces

Prefer small consumer-owned interfaces. Do not create interfaces solely for future speculation. Use composition over inheritance-like embedding. Use pointer receivers when mutating state or when copying the type is undesirable; stay consistent across a type's method set. Keep YAML/JSON/DB tags explicit and reviewed. Keep canonical advertising domain structs provider-neutral; store provider-specific extensions separately. Adapter interfaces must expose business capabilities rather than raw Meta endpoint shapes.