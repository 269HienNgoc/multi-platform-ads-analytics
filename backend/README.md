# Backend

Go backend for the multi-platform advertising analytics and campaign automation platform.

## Foundation currently implemented

- Provider-neutral advertising platform contract.
- Initial Meta Graph adapter boundary.
- Multi-account campaign workflow state machine.
- Generic automation rule evaluator.
- Bulk workflow REST endpoints.
- PostgreSQL schema for organizations, connections, ad accounts, pages, tracking sources, workflows, rules, metrics, batches and audit logs.
- YAML-based backend configuration.
- Zap structured logging for application lifecycle and HTTP requests.
- Unit tests for workflow transitions, rule safety gates and configuration loading.

## Configuration

Backend runtime configuration is YAML. Copy the example file and keep the real local config uncommitted:

```bash
cd backend
cp config/config.example.yaml config/config.yaml
```

Edit `config/config.yaml` for the machine/environment you are running on.

Important sections:

```yaml
server:
  address: ":8080"

database:
  dsn: "postgres://ads:ads@localhost:5432/ads?sslmode=disable"

redis:
  address: "localhost:6379"

meta:
  graph_base_url: "https://graph.facebook.com"
  graph_version: "v23.0"
  access_token: ""

logging:
  level: debug
  encoding: console
  development: true
```

For production, prefer `logging.encoding: json` and `logging.development: false`.

## Run locally

PostgreSQL and Redis should be installed/running directly on the host or supplied as external services. This project does not use Docker.

```bash
cd backend
go mod download
go test ./...
go run ./cmd/api -config ./config/config.yaml
```

API defaults to `http://localhost:8080` when the example configuration is used.

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
