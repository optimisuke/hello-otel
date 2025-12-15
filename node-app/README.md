# Node Todo API (Express + TypeScript + Prisma)

- Shared PostgreSQL with the FastAPI service (`todos` table)
- OpenTelemetry auto-instrumentation via `@opentelemetry/sdk-node`
- No Prisma migrations; only client generation is needed

## Local dev

```bash
cp .env.example .env
npm install
npm run prisma:generate
PORT=3001 npm run dev
```

The service exposes the same endpoints as the Python version under `http://localhost:3001/api/v1/todos` and reports telemetry to the OTLP collector (`collector:4317`).
