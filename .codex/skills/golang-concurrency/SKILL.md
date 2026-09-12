---
name: golang-concurrency
description: Goroutines, worker pools, channels, synchronization, rate limits, and concurrent campaign jobs.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Concurrency

Every goroutine must have an owner, cancellation path, and bounded lifetime. Use worker pools/semaphores for bulk account processing rather than spawning unbounded goroutines. Use `sync.WaitGroup`/`errgroup` for coordinated work and propagate context cancellation. Protect shared maps/state or avoid sharing through ownership. Respect provider rate limits and implement bounded retry with jitter. Each ad account workflow remains independent: one account failure must not cancel unrelated accounts unless the parent batch explicitly requests fail-fast behavior. Run race tests for concurrency-sensitive packages.