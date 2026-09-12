# Backend

Go API for multi-platform advertising analytics. The backend uses Clean Architecture, manual constructor injection, Gin, GORM/PostgreSQL, Viper/YAML, and Zap.

## Requirements

- Go 1.27 or newer.
- A reachable PostgreSQL instance.
- No Docker is required.

## Configuration

Copy `.env.example` to `.env` for local development and update the database values. Non-secret defaults live in `configs/config.yaml`. Environment variables use the `ADS_` prefix and override YAML values, for example `ADS_DATABASE_HOST` and `ADS_DATABASE_PASSWORD`.

VPS credentials must be supplied through environment variables or the process manager. Do not commit passwords to YAML.

## Run

```bash
make run
```

The default endpoints are:

- `GET /health/live` — process liveness.
- `GET /health/ready` — PostgreSQL readiness.

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
- `internal/config` and `internal/logging` provide centralized infrastructure configuration.
