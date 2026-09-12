# Multi-Platform Ads Analytics & Automation

Canonical advertising analytics and campaign-automation platform for synchronizing, normalizing, managing and analyzing advertising activity across Meta Ads, TikTok Ads, Google Ads and future platforms.

## Current foundation

The first campaign-automation slice now contains:

- Go REST API with multi-account bulk workflow creation.
- Provider-neutral advertising platform contract and initial Meta adapter boundary.
- Seed -> main campaign state machine.
- Rule evaluator with minimum spend/sample safety gates.
- PostgreSQL schema for accounts, pages, tracking, workflows, rules, metrics, batches and audits.
- Next.js operator dashboard for bulk account workflow creation and state monitoring.
- Docker Compose development environment with PostgreSQL and Redis.
- Architecture/ERD documentation.

## Repository layout

- `backend/` — Go backend services, APIs, workers/connectors and database migrations.
- `frontend/` — Next.js automation/analytics dashboard.
- `docs/` — architecture, canonical model, API references and ADRs.
- `.codex/skills/` — project-specific AI engineering skills.
- `deploy/` — Docker/deployment assets.
- `scripts/` — development and operational scripts.

## Local development

Backend:

```bash
cd backend
go test ./...
go run ./cmd/api
```

Frontend:

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

Or start the development stack:

```bash
cd deploy
docker compose up --build
```

Dashboard: `http://localhost:3000`  
API: `http://localhost:8080`

## Branch strategy

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

Feature development should not be committed directly to `main`.
