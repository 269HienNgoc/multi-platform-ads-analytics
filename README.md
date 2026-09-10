# Multi-Platform Ads Analytics

Canonical advertising analytics platform for synchronizing, normalizing, and analyzing advertising data across Meta Ads, TikTok Ads, Google Ads, and future platforms.

## Core architecture principles

- Normalized core data model.
- Historical configuration and performance data.
- Raw API payload retention.
- Platform-specific JSONB extensions.
- Business and CRM context.
- AI-ready analytics layer exposed through REST API and MCP.

## Repository layout

- `backend/` — Go backend services, APIs, workers, connectors, and data access.
- `frontend/` — analytics dashboard frontend.
- `docs/` — architecture, canonical model, API references, and ADRs.
- `.codex/skills/` — project-specific AI engineering skills.
- `deploy/` — deployment and infrastructure assets.
- `scripts/` — development and operational scripts.

## Branch strategy

### `main`

Production branch. Only code that has passed validation in `dev` should be merged into `main`. Production deployment is sourced from this branch.

### `dev`

Integration and localhost testing branch. Feature and fix branches merge into `dev` first. After local/integration testing succeeds, `dev` is merged into `main`.

Recommended flow:

```text
feature/* or fix/*
        ↓
       dev
        ↓
 localhost / integration testing
        ↓
       main
        ↓
   production
```
