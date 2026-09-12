# Deploy

This project intentionally does not use Docker.

Deployment assets in this directory should target native process/service management and infrastructure configuration, for example:

- systemd service units for the Go API and future workers.
- Nginx or another reverse proxy in front of the API/frontend.
- Native PostgreSQL/Redis services or managed external instances.
- Environment-specific YAML configuration stored outside Git and passed with `-config`.
- CI/CD scripts that build Go binaries and Next.js artifacts directly.

Do not commit production credentials or a real `backend/config/config.yaml` file.
