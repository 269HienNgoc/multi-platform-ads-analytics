---
name: golang-naming
description: Go naming conventions for packages, APIs, interfaces, errors, tests, and domain entities.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go"]
---

# Naming

Use short lowercase package names without underscores. Use MixedCaps identifiers and conventional initialisms (`ID`, `URL`, `HTTP`, `API`). Constructors use `NewX`. Avoid generic `util`, `helper`, `manager` packages when a domain name is available. Interfaces should describe behavior (`CampaignRepository`, `AdPlatform`) and usually live near the consumer. Error variables use `ErrX`. Boolean names should read naturally (`enabled`, `isActive`, `hasPermission`). Test names should state behavior and scenario.