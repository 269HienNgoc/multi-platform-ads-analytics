# Backend

Go API for multi-platform advertising analytics. The backend uses Clean Architecture, manual constructor injection, Gin, GORM/PostgreSQL, Viper/YAML, and Zap.

## Requirements

- Go 1.27 or newer.
- A reachable PostgreSQL instance.
- No Docker is required.

## Configuration

Copy `.env.example` to `.env` for local development and update the database values. Non-secret defaults live in `configs/config.yaml`. Environment variables use the `ADS_` prefix and override YAML values, for example `ADS_DATABASE_HOST` and `ADS_DATABASE_PASSWORD`.

VPS credentials must be supplied through environment variables or the process manager. Do not commit passwords to YAML.

## Run locally

Clone the repository and switch to the integration branch:

```bash
git clone https://github.com/269HienNgoc/multi-platform-ads-analytics.git
cd multi-platform-ads-analytics
git switch dev
cd backend
```

Create the local database once, then copy and update the environment file:

```bash
createdb -U postgres multi_platform_ads
cp .env.example .env
```

Apply the versioned schema before starting the API:

```bash
make migrate-up
make run
```

If `make` is not available, including on a default Windows installation, use the equivalent Go commands:

```bash
go run ./cmd/migrate -config configs/config.yaml -migrations migrations -action up
go run ./cmd/api -config configs/config.yaml
```

The default endpoints are:

- `GET /health/live` — process liveness.
- `GET /health/ready` — PostgreSQL readiness.
- `GET /api/v1/ad-accounts` — list normalized accounts and campaign counts.
- `POST /api/v1/ad-accounts` — create an advertising account.
- `POST /api/v1/campaigns` — create a campaign.
- `POST /api/v1/ad-groups` — create an ad group/ad set.
- `POST /api/v1/ads` — create an ad.
- `POST /api/v1/creatives` — create a creative.
- `GET /api/v1/ad-accounts/{accountID}/hierarchy` — read the complete account hierarchy.
- `GET /api/v1/workflows` — list campaign automation workflows.
- `POST /api/v1/workflows/bulk` — create one workflow per selected account.
- `POST /api/v1/workflows/{workflowID}/transition` — apply a valid state transition.
- `POST /api/v1/workflows/{workflowID}/metrics` — record metrics and advance a completed seed campaign.
- `GET /api/v1/connectors/meta` — read Meta connector status.
- `POST /api/v1/connectors/meta/sync` — run a read-only Meta account/campaign sync.
- `GET /api/v1/connectors/meta/sync-runs` — list recent Meta sync attempts.

Example account request:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ad-accounts \
  -H "Content-Type: application/json" \
  -d '{"platform":"meta","external_id":"act_123","name":"Local test","currency":"USD","timezone":"Asia/Ho_Chi_Minh","status":"active","provider_data":{}}'
```

Supported canonical platforms are `meta`, `tiktok`, and `google`. Supported entity statuses are `active`, `paused`, and `archived`.

Create a workflow after creating an account:

```bash
curl -X POST http://127.0.0.1:8080/api/v1/workflows/bulk \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: local-test-001" \
  -d '{"organization_id":"local-testing","ad_account_ids":["ACCOUNT_UUID"],"page_external_id":"PAGE_ID","seed_spend_limit_usd":10}'
```

Workflow state and metrics are durable in PostgreSQL. Provider publish and budget operations remain fail-closed until validation and approval guardrails are implemented.

In production, configure `ADS_SERVER_API_KEY` and send it as `X-API-Key` for every `/api/v1/*` request. To enable the read-only Meta connector, set `ADS_META_ENABLED=true`, `ADS_META_ACCESS_TOKEN`, and the required version/base URL values. Never commit tokens.

## Database migrations

Migrations are stored in `migrations/` and applied transactionally. The service never calls GORM `AutoMigrate`.

```bash
make migrate-version
make migrate-up
make migrate-down # rolls back one migration and can delete data
```

Run only one migration process at a time during deployment. Database passwords belong in `.env` locally or VPS environment variables in production.

## Validate

```bash
make check
```

## Package boundaries

- `cmd/api` contains only the executable entry point.
- `internal/app` is the composition root and owns lifecycle wiring.
- `internal/domain` contains framework-free advertising concepts.
- `internal/application` contains use cases and the interfaces they consume.
- `internal/adapter` contains Gin and PostgreSQL implementations.
- `internal/adapter/meta` contains the guarded Meta Graph API connector foundation.
- `internal/config` and `internal/logging` provide centralized infrastructure configuration.
- `migrations` contains reviewed PostgreSQL schema changes and rollback files.
