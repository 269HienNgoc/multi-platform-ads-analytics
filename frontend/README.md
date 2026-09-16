# Frontend dashboard

Next.js dashboard for operating advertising accounts across Meta, TikTok, and Google.

## Requirements

- Node.js 20.9 or newer
- npm 10 or newer
- Go backend running at `http://127.0.0.1:8080`

## Run locally

```bash
npm install
npm run dev
```

Open `http://localhost:3000`. Next.js forwards `/backend/*` requests to the Go API so local development does not require CORS.

To use another backend address, create `.env.local`:

```bash
BACKEND_API_URL=http://127.0.0.1:8080
BACKEND_API_KEY=the-same-value-as-ADS_SERVER_API_KEY
```

Production additionally requires `FRONTEND_AUTH_USER` and `FRONTEND_AUTH_PASSWORD`. Serve production over HTTPS because the dashboard uses HTTP Basic authentication as its current operator-access layer. Backend credentials are attached only by the server-side Route Handler and are never returned to browser JavaScript.

## Current integration

- `/health/ready` drives the backend connection indicator.
- `GET /api/v1/ad-accounts` supplies the real account list and campaign counts.
- The “Đồng bộ Meta” action calls the read-only Meta connector, then refreshes the account list.
- The “Thêm tài khoản” form calls `POST /api/v1/ad-accounts`.
- Bulk workflow creation sends an idempotency key and displays partial account failures without discarding successful accounts.
- KPI, chart, recommendations, and spend/ROAS values remain explicitly marked as simulated until metrics read endpoints are implemented.

See `../docs/api-reference/backend-api-report.md` for the current API coverage and frontend gaps.
