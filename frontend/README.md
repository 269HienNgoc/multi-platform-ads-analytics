# Frontend

Next.js dashboard for multi-account campaign automation and analytics.

## Local development

```bash
cp .env.example .env.local
npm install
npm run dev
```

Open `http://localhost:3000` while the Go API is running on `http://localhost:8080`.

The first dashboard slice includes:

- Multi-account bulk workflow creation.
- Per-account workflow state table.
- Seed spend-limit configuration.
- Pixel/event/post fields that will later be populated from synced platform assets.
