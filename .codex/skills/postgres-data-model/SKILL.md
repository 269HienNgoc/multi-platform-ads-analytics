---
name: postgres-data-model
description: Design PostgreSQL schemas, GORM persistence models, migrations, indexing, partitioning, JSONB, and historical data structures for advertising data in this project.
---

# postgres-data-model

> This skill supersedes `samber/cc-skills-golang@golang-database` where that skill prohibits ORM usage. This project has explicitly selected GORM.

Use GORM for PostgreSQL access and keep GORM models inside PostgreSQL adapter packages. Domain entities must not contain GORM tags or import persistence libraries. Do not use `AutoMigrate` in application startup; schema changes belong in reviewed, versioned migrations. Use GORM's context-aware APIs and configure its underlying connection pool only inside the PostgreSQL adapter.

Keep credentials out of YAML and source control. Load database settings through the central Viper configuration package and use environment overrides for secrets on the VPS.

Follow `AGENTS.md` and the canonical multi-platform architecture. Do not introduce platform-specific assumptions into normalized core models unless explicitly documented.
