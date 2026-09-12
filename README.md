# Multi-Platform Ads Analytics & Automation

Canonical advertising analytics and campaign-automation platform for synchronizing, normalizing, managing and analyzing advertising activity across Meta Ads, TikTok Ads, Google Ads and future platforms.

## Current foundation

The first campaign-automation slice now contains:

- Go REST API with multi-account bulk workflow creation.
- Provider-neutral advertising platform contract and initial Meta adapter boundary.
- Seed -> main campaign state machine.
- Rule evaluator with minimum spend/sample safety gates.
- PostgreSQL schema for accounts, pages, tracking, workflows, rules, metrics, batches and audits.
- YAML-based backend configuration.
- Zap structured logging.
- Next.js operator dashboard for bulk account workflow creation and state monitoring.
- Architecture/ERD documentation.

## Runtime conventions

- Backend: Go.
- Frontend: Next.js.
- System of record: PostgreSQL.
- Queue/cache coordination: Redis.
- Backend configuration: YAML files.
- Backend logging: Zap.
- Docker is intentionally not used by this project.

## Repository layout

- `backend/` — Go backend services, APIs, workers/connectors and database migrations.
- `frontend/` — Next.js automation/analytics dashboard.
- `docs/` — architecture, canonical model, API references and ADRs.
- `.codex/skills/` — project-specific AI engineering skills.
- `deploy/` — native deployment/infrastructure notes and service-manager configuration.
- `scripts/` — development and operational scripts.

## Local development

Backend:

```bash
cd backend
cp config/config.example.yaml config/config.yaml
# Edit config/config.yaml for local PostgreSQL, Redis and Meta settings.
go mod download
go test ./...
go run ./cmd/api -config ./config/config.yaml
```

Frontend:

```bash
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

PostgreSQL and Redis must be available locally or through external services configured in `backend/config/config.yaml`.

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
