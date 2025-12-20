# Quarkus Todo API

Quarkus (RESTEasy Reactive + Hibernate ORM) 版 Todo API です。既存の PostgreSQL `todos` テーブルをそのまま利用します（マイグレーション不要）。Java 21 で動作確認しています。

## ローカル開発

```bash
cd quarkus-app
cp .env.example .env

# 依存解決とビルド（JVMランナー）
mvn -B -DskipTests package

# アプリ起動（JVMランナー）
PORT=8081 DATABASE_JDBC_URL=jdbc:postgresql://localhost:5432/tododb \
DATABASE_USERNAME=todouser DATABASE_PASSWORD=todopass \
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
OTEL_SERVICE_NAME=todo-api-quarkus \
java -jar target/*-runner.jar

# 開発モード（ホットリロード）
mvn quarkus:dev
# 保存ごとに再コンパイル・再デプロイされます。ポート変更は -Dquarkus.http.port=8081 などで上書き。

# コンテナイメージを Quarkus 拡張でビルド（Docker 必須）
# ネイティブ版（推奨、compose もこちらを参照）
mvn -DskipTests \
  -Dquarkus.package.type=native \
  -Dquarkus.native.container-build=true \
  -Dquarkus.container-image.build=true \
  -Dquarkus.container-image.builder=docker \
  -Dquarkus.container-image.registry=localhost \
  -Dquarkus.container-image.group=todogroup \
  -Dquarkus.container-image.name=todo-api-quarkus \
  -Dquarkus.container-image.tag=native \
  package

# docker-compose から起動（ネイティブ版イメージ作成後）
docker-compose up -d quarkus-api

# テスト
# 単体/統合テスト（RestAssured + Mockito モック、DBアクセスはリポジトリをモック）
mvn test
```

## エンドポイント

- `GET /health`
- `GET /api/v1/todos?skip=0&limit=100`
- `GET /api/v1/todos/{id}`
- `POST /api/v1/todos`
- `PUT /api/v1/todos/{id}`
- `DELETE /api/v1/todos/{id}`

リクエスト/レスポンスのキーは `snake_case`（例: `created_at`）です。

## コード概要（パッケージ）

- `model/TodoEntity`: `todos` テーブルをマッピングした Hibernate/Panache エンティティ（UUID 主キー、snake_case カラム対応）。
- `repository/TodoRepository`: PanacheRepositoryBase でクエリを集約（作成日時の降順でページング取得）。
- `service/TodoService`: CRUD のドメインロジック。存在しない ID で 404 を返し、null の completed を false として作成。
- `dto/*`: リクエスト/レスポンス DTO（snake_case にシリアライズ）。
- `api/TodoResource`: `/api/v1/todos` の REST エンドポイント。入力バリデーション、400/404/201/204 をハンドリング。
- `api/HealthResource`: `/health` の疎通確認。
- `exception/*Mapper`: バリデーションや不正リクエスト時のレスポンス整形。

### Quarkus の動き（エントリ/ルーティング/DI）

- `mvn package` 時に Quarkus が `*-runner.jar` にエントリポイント (`io.quarkus.runner.GeneratedMain`) を自動生成。`java -jar ...-runner.jar` でそこから起動。
- ルーティングは JAX-RS アノテーションで定義（`@Path`/`@GET`/`@POST` など）。別途ルータークラスは不要。
- DI は CDI (Arc) が `@Inject` で配線。Resource → Service → Repository の依存はコード上のアノテーションだけで解決。
- 設定は `application.properties` に集約 (`quarkus.http.*`, `quarkus.datasource.*`, `quarkus.otel.*` など)。

### Dockerfile の使い分け

- `Dockerfile`（ルート）: 手書きのマルチステージ（JVMランナーをビルドしてコピー）用。Quarkus 拡張を使わず自前ビルドしたい場合に利用。
- `src/main/docker/Dockerfile.jvm`: Quarkus container-image-docker 拡張のデフォルト（JVMランナー前提）。`quarkus.package.type=jar` のときに使われる。
- `src/main/docker/Dockerfile.native`: ネイティブバイナリ前提。`quarkus.package.type=native` で `quarkus.native.container-build=true` を指定した際に使われ、`localhost/todogroup/todo-api-quarkus:native` を生成する。
