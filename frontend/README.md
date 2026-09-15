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
```

## Current integration

- `/health/ready` drives the backend connection indicator.
- The “Thêm tài khoản” form calls `POST /api/v1/ad-accounts`.
- KPI, chart, account list, and workflow data are explicitly marked as demo data because the current backend does not expose list or metrics read endpoints yet.

See `../docs/api-reference/backend-api-report.md` for the current API coverage and frontend gaps.
