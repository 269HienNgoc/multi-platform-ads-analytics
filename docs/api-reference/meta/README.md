# Meta connector notes

The implemented connector is read-only for catalog synchronization:

- Cursor-paginates `/me/adaccounts` and maps account identity, currency, timezone, and lifecycle status.
- Cursor-paginates `/{ad-account-id}/campaigns` for every visible account.
- Reads Pages and Pixels as optional assets. Missing optional permissions produce sync warnings rather than discarding account/campaign data.
- Retries `429`, `500`, `502`, `503`, and `504` responses up to three attempts, honoring `Retry-After` with a 30-second cap.
- Persists normalized accounts/campaigns, archives provider-managed accounts or campaigns no longer returned by a successful full read, leaves manually created records untouched, retains raw payloads, and records every sync run.

Configure the Graph API version explicitly with `meta.version` in the backend YAML; do not assume the repository value remains valid indefinitely. Put the token in `meta.access_token` in a protected deployment YAML. Tokens are server-only and must never be returned by an API response or exposed through `NEXT_PUBLIC_*` variables.

Campaign publishing, budget mutation, and Insights attribution remain fail-closed. They require reviewed provider-version payloads, preflight checks, human approval, idempotency, budget guardrails, and audit events before they can be enabled.
