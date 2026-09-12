# Architecture

The backend follows Clean Architecture with manual constructor injection. Dependencies point inward: adapters depend on application ports and domain types; the domain never imports frameworks or infrastructure packages.

```text
backend/
├── cmd/api/                    # Minimal executable entry point
├── configs/                    # Non-secret YAML defaults
├── internal/
│   ├── app/                    # Composition root and lifecycle
│   ├── domain/ads/             # Canonical, provider-neutral concepts
│   ├── application/health/     # Use case and consumed port
│   ├── adapter/httpapi/        # Gin routes and middleware
│   ├── adapter/postgres/       # GORM/PostgreSQL implementation
│   ├── config/                 # Viper loading and validation
│   └── logging/                # Zap construction and flushing
├── Makefile
└── go.mod
```

## Boundary rules

- Domain models never carry GORM tags and never import Gin, GORM, Viper, Zap, or provider SDKs.
- Interfaces are declared by the application package that consumes them.
- PostgreSQL records and mapping code belong in `adapter/postgres`, not in the domain.
- Meta, TikTok, and Google implementations will live in provider adapters behind canonical connector ports.
- Raw provider payloads and provider-specific fields remain outside the normalized domain and are persisted as adapter-owned JSONB data.
- API and worker binaries share application/domain packages but own independent process lifecycles.

## Configuration and deployment

YAML contains safe defaults. Environment variables override YAML through the central Viper loader, and sensitive VPS values are never committed. The service emits structured Zap logs to standard output and shuts down HTTP and database resources gracefully. Docker is intentionally not part of the deployment model.
