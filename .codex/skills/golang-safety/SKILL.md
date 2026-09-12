---
name: golang-safety
description: Defensive Go coding against panics, races, aliasing, nil bugs, and silent data corruption.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Safety

Treat nil maps, typed-nil interfaces, unchecked type assertions, integer narrowing, shared slices/maps, and goroutine leaks as correctness risks. Never access a map concurrently without synchronization. Copy mutable slices/maps when ownership crosses boundaries. Close resources exactly once and defer cleanup immediately after successful acquisition. Bound goroutines and worker concurrency. Guard state transitions so retries cannot create duplicate campaigns or duplicate spend actions. Prefer explicit zero-value semantics and validate external API payloads before domain use.