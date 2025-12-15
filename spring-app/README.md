# Spring Boot Todo API

Spring Boot + Spring Data JPA で実装した Todo API。既存の PostgreSQL `todos` テーブルをそのまま利用し、マイグレーションは行いません。OpenTelemetry Java Agent で自動計装します。

## クイックスタート

```bash
cp spring-app/.env.example spring-app/.env  # 必要なら
# または環境変数を docker-compose から渡す

docker-compose build spring-api
docker-compose up -d spring-api
```

- API: http://localhost:8080/api/v1/todos
- Health: http://localhost:8080/actuator/health (デフォルト無効、必要に応じて有効化)

## 主要エンドポイント
- `GET /api/v1/todos` 全件取得
- `GET /api/v1/todos/{id}` 取得
- `POST /api/v1/todos` 作成
- `PUT /api/v1/todos/{id}` 更新
- `DELETE /api/v1/todos/{id}` 削除

## OpenTelemetry Java Agent
- `-javaagent:/app/opentelemetry-javaagent.jar` を付与（Dockerfile に同梱）
- 代表的な環境変数:
  - `OTEL_EXPORTER_OTLP_ENDPOINT=http://collector:4317`
  - `OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://collector:4318/v1/metrics`
  - `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT=http://collector:4318/v1/logs`
  - `OTEL_SERVICE_NAME=todo-api-spring`

## ビルド/ローカル実行

```bash
cd spring-app
mvn clean package -DskipTests
java -javaagent:./opentelemetry-javaagent.jar -jar target/todo-api-spring-0.1.0.jar
```

## 注意
- 既存の `todos` テーブルを前提としており、`spring.jpa.hibernate.ddl-auto=none` でマイグレーションを行いません。
- created_at / updated_at はアプリ側で現在時刻を設定します。
