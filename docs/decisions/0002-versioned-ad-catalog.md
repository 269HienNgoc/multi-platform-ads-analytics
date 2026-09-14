# ADR-0002: Versioned advertising catalog

- Status: Accepted
- Date: 2026-09-14

## Context

The first executable backend foundation had database connectivity and health checks but no durable advertising hierarchy. Local and VPS deployments also needed a repeatable schema lifecycle without GORM `AutoMigrate`.

## Decision

- Store accounts, campaigns, ad groups, ads, and creatives in separate normalized PostgreSQL tables.
- Use application-generated UUIDs as internal identifiers and keep provider identifiers in scoped `external_id` columns.
- Keep provider-specific fields in JSONB while preserving normalized cross-platform fields.
- Add daily performance, sync-run, and raw-payload tables now so subsequent connector work has stable historical storage.
- Apply paired `up.sql` and `down.sql` files transactionally through a dedicated one-off command.
- Expose initial create endpoints and a complete account hierarchy read endpoint under `/api/v1`.

## Consequences

- Provider connectors can upsert into a consistent hierarchy without leaking SDK types into the domain.
- Schema changes are explicit and reviewable; application startup never mutates schema.
- The initial API is suitable for localhost development only until authentication and tenant authorization are implemented.
