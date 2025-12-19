# Quarkus Todo API

Quarkus (RESTEasy Reactive + Hibernate ORM) 版 Todo API です。既存の PostgreSQL `todos` テーブルをそのまま利用します（マイグレーション不要）。Java 21 で動作確認しています。

## ローカル開発

```bash
cd quarkus-app
cp .env.example .env

# 依存解決とビルド（Java 21）
mvn -B -DskipTests package

# アプリ起動
PORT=8081 DATABASE_JDBC_URL=jdbc:postgresql://localhost:5432/tododb \
DATABASE_USERNAME=todouser DATABASE_PASSWORD=todopass \
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
OTEL_SERVICE_NAME=todo-api-quarkus \
java -jar target/*-runner.jar
```

## エンドポイント

- `GET /health`
- `GET /api/v1/todos?skip=0&limit=100`
- `GET /api/v1/todos/{id}`
- `POST /api/v1/todos`
- `PUT /api/v1/todos/{id}`
- `DELETE /api/v1/todos/{id}`

リクエスト/レスポンスのキーは `snake_case`（例: `created_at`）です。

## コード概要

- `TodoEntity`: `todos` テーブルをマッピングした Hibernate/Panache エンティティ（UUID 主キー、snake_case カラム対応）。
- `TodoRepository`: PanacheRepositoryBase でクエリを集約（作成日時の降順でページング取得）。
- `TodoService`: CRUD のドメインロジック。存在しない ID で 404 を返し、null の completed を false として作成。
- `TodoResource`: REST エンドポイント（`/api/v1/todos`）。入力バリデーション、400/404/201/204 をハンドリング。
- `ValidationExceptionMapper` / `BadRequestExceptionMapper`: バリデーション/不正リクエストのレスポンスを整形。
- `HealthResource`: `/health` で疎通確認。
