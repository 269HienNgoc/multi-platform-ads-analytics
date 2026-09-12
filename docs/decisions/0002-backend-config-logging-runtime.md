# ADR 0002: Backend configuration, logging and runtime

## Status

Accepted

## Context

The backend needs a consistent runtime convention before Meta credentials, PostgreSQL repositories and background workers are added. The project owner does not want Docker in the development/deployment workflow and prefers explicit file-based backend configuration plus structured logging.

## Decision

1. Run backend services as native Go processes rather than Docker containers.
2. Use YAML as the backend runtime configuration format.
3. Pass the configuration path explicitly with `-config`; the default local path is `./config/config.yaml`.
4. Keep `config.example.yaml` in Git, but ignore real local/environment configuration files that may contain credentials.
5. Use Zap for backend application and HTTP request logging.
6. Use console logging for local development when desired and JSON structured logging for production.
7. Keep PostgreSQL and Redis as native/local or managed external services; connection settings come from YAML.

## Consequences

- Local development does not require Docker or Docker Compose.
- Configuration is reviewable and grouped by subsystem instead of being spread across backend environment variables.
- Secrets must be protected through filesystem/deployment practices because real YAML config files are not committed.
- Native deployment should use a process manager such as systemd and a reverse proxy where appropriate.
- Logging becomes machine-readable and suitable for later aggregation/observability pipelines.
