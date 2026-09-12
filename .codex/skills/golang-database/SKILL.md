---
name: golang-database
description: PostgreSQL access, repositories, transactions, migrations, pooling, and query safety in Go.
license: MIT
metadata:
  source_repo: https://github.com/samber/cc-skills-golang
  source_commit: 19a0626ae8565d27a7b7bdf59d8d99d94d7e284c
paths: ["backend/**/*.go", "backend/migrations/**/*.sql"]
---

# Database

PostgreSQL is the system of record. Use repositories at domain boundaries; keep SQL/provider persistence concerns out of HTTP handlers. Parameterize all SQL. Use transactions for state changes that must be atomic, especially workflow transition + audit/outbox records. Configure a bounded connection pool and pass contexts to queries. External platform IDs are unique business keys, not internal primary keys. Migrations must be deterministic and reversible where practical. Use JSONB only for provider-specific/raw payload extensions; normalize fields used for joins, filters, rules, and analytics. Avoid N+1 queries in account/campaign dashboards.