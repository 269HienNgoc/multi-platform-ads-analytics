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

Backend runtime configuration is YAML. Copy the example file and keep the real config uncommitted:

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
  dsn: "postgres://DB_USER:DB_PASSWORD@127.0.0.1:5432/DB_NAME?sslmode=disable"

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

### PostgreSQL on the VPS

When the Go backend and PostgreSQL run on the same VPS, keep PostgreSQL bound to the VPS itself and connect through loopback:

```yaml
database:
  dsn: "postgres://YOUR_DB_USER:YOUR_DB_PASSWORD@127.0.0.1:5432/YOUR_DB_NAME?sslmode=disable"
```

Replace `YOUR_DB_USER`, `YOUR_DB_PASSWORD` and `YOUR_DB_NAME` with the PostgreSQL account and database created on the VPS. The real `config/config.yaml` is ignored by Git and must not be committed because it can contain database passwords and platform access tokens.

If PostgreSQL is hosted on a different server or managed database service, replace `127.0.0.1` with its private IP/DNS name and use the SSL mode required by that database provider.

For production, prefer `logging.encoding: json` and `logging.development: false`.

## Run locally or on the VPS

PostgreSQL and Redis should be installed/running directly on the host or supplied as external services. This project does not use Docker.

```bash
cd backend
go mod download
go test ./...
go run ./cmd/api -config ./config/config.yaml
```

API defaults to port `8080` when the example configuration is used.

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
