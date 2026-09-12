---
name: golang-design-patterns
description: Idiomatic Go architecture patterns for adapters, repositories, workflows, middleware, and resilience.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Design patterns

Prefer constructor injection and explicit dependencies. Use adapter/port boundaries for Meta, TikTok, Google, storage, and queues. Use functional options only when configuration truly has optional dimensions; otherwise explicit config structs are clearer. Model campaign automation as explicit state transitions rather than hidden side effects. Make commands idempotent with stable idempotency keys. Use middleware for transport-level concerns such as request logging and auth, not domain business logic. Implement circuit/retry behavior at outbound integration boundaries.