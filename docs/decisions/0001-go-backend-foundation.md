# ADR-0001: Go backend foundation

- Status: Accepted
- Date: 2026-09-12

## Context

The product must synchronize and analyze advertising data from multiple providers without coupling its canonical model to Meta, TikTok, Google, or any future provider. It also needs independently evolvable HTTP, persistence, sync-worker, and analytics boundaries.

## Decision

- Use Clean Architecture organized around `domain`, `application`, and `adapter` boundaries.
- Use manual constructor injection while the dependency graph remains small.
- Keep process composition and graceful lifecycle management in `internal/app`; keep `cmd` entry points minimal.
- Use Gin only in the HTTP adapter.
- Use GORM only in PostgreSQL adapter packages. Domain entities do not contain GORM tags.
- Use reviewed, versioned migrations rather than `AutoMigrate` during application startup.
- Use Viper as the only configuration gateway. YAML provides non-secret defaults and `ADS_` environment variables provide deployment overrides.
- Use Zap for all structured application logs.
- Run without Docker; connect to the PostgreSQL instance configured on the target VPS.

## Consequences

- Provider and transport implementations can be replaced without changing the domain.
- Persistence records require explicit mapping to and from domain entities.
- Initial wiring is intentionally explicit; a DI framework should be considered only when manual wiring becomes difficult to maintain.
- Configuration and logging conventions are enforceable from one package rather than duplicated throughout the codebase.
- The API can add a separate worker binary later without sharing process lifecycle state.
