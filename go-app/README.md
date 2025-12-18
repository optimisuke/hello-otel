# Go Todo API (chi + sqlx + pgx)

PostgreSQL 共有の Todo API を Go で実装したものです。Loongsuite Go Agent でビルド時に自動計装（トレース/メトリクス）し、アプリ側のログは zap の構造化ログを stdout とファイルに出します（OTEL SDK は使いません）。ログの OTEL 送信は Collector の Filelog Receiver で行います。

## ローカル実行

```bash
cd go-app
cp .env.example .env

# 依存取得＆ビルド/テスト
GOTOOLCHAIN=auto go test ./...

# 実行（OTEL送信先は collector:4317 を想定）
PORT=3002 DATABASE_URL=postgresql://todouser:todopass@localhost:5432/tododb \
  LOG_FILE_PATH=/tmp/go-api.log \
  OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
  OTEL_EXPORTER_OTLP_PROTOCOL=grpc \
  OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://localhost:4318/v1/metrics \
  OTEL_EXPORTER_OTLP_METRICS_PROTOCOL=http/protobuf \
  OTEL_SERVICE_NAME=todo-api-go \
  go run ./cmd/server
```

## エンドポイント

- `GET    /health`
- `GET    /api/v1/todos?skip=0&limit=100`
- `GET    /api/v1/todos/{id}`
- `POST   /api/v1/todos`
- `PUT    /api/v1/todos/{id}`
- `DELETE /api/v1/todos/{id}`

レスポンスフィールドは他言語実装と同じ `id, title, description, completed, created_at, updated_at` です。

## Docker ビルド (単体)

```bash
cd go-app
docker build -t todo-api-go .
```

## OTEL 自動計装（Loongsuite）

- Docker ビルド時に `otel go build` を利用し、自動計装されたバイナリを生成しています（トレース/メトリクス）。アプリコードでは OTEL SDK を直接使っていません。
- 送信先は `OTEL_EXPORTER_OTLP_ENDPOINT`（gRPC）など標準 OTEL 環境変数で指定します。compose では `collector:4317` に送信する設定済みです。

## ログ（stdout → filelog）

- compose では Go API が `LOG_FILE_PATH=/var/log/go-api/app.log` にもログを書き出し、Collector の Filelog Receiver がそれを読み取って Loki に送ります。
