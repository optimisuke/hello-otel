# Node Todo API (Express + TypeScript + Prisma)

FastAPI 版と同じ PostgreSQL の `todos` テーブルを共有する Node 実装です。OpenTelemetry は `@opentelemetry/sdk-node` の自動計装を使用し、Prisma はクライアント生成のみ（マイグレーション不要）です。

## ローカル開発

```bash
cp .env.example .env
npm install
npm run prisma:generate
PORT=3001 npm run dev
```

エンドポイントは Python 版と同じく `http://localhost:3001/api/v1/todos` 配下にあり、テレメトリは OTLP (collector:4317/4318) に送信されます。
