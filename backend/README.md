# Backend

Go backend for the multi-platform advertising analytics and campaign automation platform.

## What is implemented in this foundation

- Provider-neutral advertising platform contract.
- Initial Meta Graph adapter boundary.
- Multi-account campaign workflow state machine.
- Generic automation rule evaluator.
- Bulk workflow REST endpoints.
- PostgreSQL schema for organizations, connections, ad accounts, pages, tracking sources, workflows, rules, metrics, batches and audit logs.
- Unit tests for workflow transitions and rule safety gates.

## Run locally

```bash
cp .env.example .env
go test ./...
go run ./cmd/api
```

API defaults to `http://localhost:8080`.

### Useful endpoints

```text
GET  /health
GET  /api/v1/workflows
POST /api/v1/workflows/bulk
POST /api/v1/workflows/{id}/transition
POST /api/v1/workflows/{id}/metrics
```

Example bulk request:

```json
{
  "organizationId": "org-demo",
  "adAccountIds": ["act_001", "act_002", "act_003"],
  "pageId": "page_001",
  "pixelId": "pixel_001",
  "pixelEvent": "CompleteRegistration",
  "existingPostId": "post_001",
  "seedSpendLimitUsd": 10
}
```

The current service uses an in-memory workflow store so the state-machine API can be validated immediately. The SQL migration defines the durable PostgreSQL model that replaces the in-memory repository in the next integration step.
