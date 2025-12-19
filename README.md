# Todo API with OpenTelemetry & Grafana OTEL-LGTM

最もシンプルな構成で完全な観測性を実現した Todo アプリケーション。
Python(FastAPI)、Node(Express)、Java(Spring Boot)、Go(chi)、Rust(axum) の 5 実装を同梱し、同じ PostgreSQL に接続します。

## ✨ 特徴

- 🚀 **観測基盤は 4 サービス** - app, postgres, lgtm（統合観測基盤）, otel-collector（spanmetrics）に加え Node/Go/Rust 版 API サービスも同梱
- 🎯 **設定ファイル不要** - docker-compose.yml のみ
- 📊 **完全な観測性** - トレース + ログ + メトリクス
- 🧹 **クリーンコード** - アプリに観測性コードゼロ
- ⚡ **すぐ使える** - 起動後即座に Grafana で確認可能
- 🧩 **Node/Go/Rust 版も同梱** - Express + TypeScript + Prisma、chi + sqlx、axum + sqlx で同じ PostgreSQL を共有
- 🔧 **uv 管理** - 高速な依存関係管理

## 🛠 技術スタック

### アプリケーション（5 実装）

- **Python**: FastAPI + SQLAlchemy + Alembic（`python-app/`）
- **Node.js**: Express + TypeScript + Prisma（`node-app/`）
- **Java**: Spring Boot + Spring Data JPA（`spring-app/`、Java Agentで自動計装）
- **Go**: chi + sqlx + pgx（`go-app/`）
- **Rust**: axum + sqlx（`rust-app/`、Beyla(eBPF) で自動計装）
- **PostgreSQL**: 共通データベース

### 観測性（LGTM 統合）

- **Grafana OTEL-LGTM**: Grafana / Tempo / Loki / Mimir
- **OpenTelemetry Collector**: spanmetrics connector 付き

### 自動計装

- Python: `opentelemetry-instrument`
- Node: `@opentelemetry/sdk-node` + auto-instrumentations
- Java: OpenTelemetry Java Agent（`-javaagent`）
- Go: Loongsuite（`otel go build`）
- Rust: Grafana Beyla（eBPF / Linux カーネル要件あり）

## 📦 前提条件

- Docker Desktop 4.0+
- Docker Compose 2.0+

## 🚀 クイックスタート

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd hello-otel
```

### 2. 環境変数の設定（オプション）

```bash
cp python-app/.env.example python-app/.env
# Node 版も動かす場合はこちらもコピー
cp node-app/.env.example node-app/.env
# Spring 版も動かす場合はこちらもコピー
cp spring-app/.env.example spring-app/.env
# Go 版も動かす場合はこちらもコピー
cp go-app/.env.example go-app/.env
# 必要に応じて .env を編集
```

### 3. サービスの起動

```bash
docker-compose up -d
docker-compose run --rm --profile testing k6   # k6 負荷テスト（任意）
```

### 4. データベースマイグレーション

FastAPI 版のマイグレーション手順は `python-app/README.md` を参照してください。
Node 版（Express）は同じ `todos` テーブルを利用するため、Prisma のマイグレーションは不要です（クライアント生成のみ）。

### 5. アクセス

| サービス     | URL                        | 説明                   |
| ------------ | -------------------------- | ---------------------- |
| **API (FastAPI)** | http://localhost:8000      | Python 版 Todo API     |
| **API Docs** | http://localhost:8000/docs | Swagger UI             |
| **API (Node)** | http://localhost:3001      | Express + TypeScript 版 Todo API |
| **API (Spring)** | http://localhost:8080      | Spring Boot 版 Todo API |
| **API (Quarkus)** | http://localhost:8081      | Quarkus 版 Todo API |
| **API (Go)** | http://localhost:3002      | chi + sqlx 版 Todo API |
| **API (Rust)** | http://localhost:3003      | axum + sqlx 版 Todo API |
| **Grafana**  | http://localhost:3000      | 統合ダッシュボード     |

**Grafana 初回ログイン**

- ユーザー名: `admin`
- パスワード: `admin`

## 📊 観測性の確認

Grafana にアクセス（http://localhost:3000）して：

### トレースの確認

1. **Explore** をクリック
2. データソース: **Tempo** を選択
3. **Search** タブでトレースを検索
4. リクエストのフローを確認

### LLM API（OpenAI）呼び出しのトレース（job コンテナ）

OpenLLMetry / Traceloop SDK を使って、OpenAI API 呼び出しのスパンを OTLP で Collector に送ります。

```bash
# 先に観測基盤だけ上げる（全部 up していてもOK）
docker-compose up -d lgtm collector

# OpenAI のAPIキーを渡して job を1回実行（--profile llm）
export OPENAI_API_KEY="sk-..."
docker-compose --profile llm run --rm llm-job
```

- 追加で指定できる環境変数: `OPENAI_MODEL` / `OPENAI_PROMPT`
- Grafana → Explore → Tempo で `service.name=llm-job` を検索すると見つけやすいです
- ログは Grafana → Explore → Loki で `{service_name="llm-job"}` を検索すると確認できます
  - ※ Traceloop のログ出力はデフォルト無効なので、compose 側で `TRACELOOP_LOGGING_ENABLED=true` を指定しています

### ログの確認

1. **Explore** をクリック
2. データソース: **Loki** を選択
3. LogQL クエリ: `{service_name="todo-api"}`
4. 直近のデータが無い場合は時間範囲を「Last 5m」に変更
5. バリデーションエラーやアクセスログなど、FastAPI が `logger.info`/`logger.warning` で出したものが流れます（`OTEL_LOGS_EXPORTER` は compose で有効化済み）

### メトリクスの確認

1. **Explore** をクリック
2. データソース: **Mimir** を選択
3. PromQL クエリ例:
   ```promql
   rate(http_server_duration_count[5m])
   ```
4. Span Metrics（トレースから集約されたメトリクス）例:

   ```promql
   # ルート別のレイテンシヒストグラム
   sum by (http_method, http_route, http_status_code, le) (rate(spanmetrics_latency_bucket[5m]))

   # ルート別のリクエストレート
   sum by (http_method, http_route, http_status_code) (rate(spanmetrics_latency_count[5m]))
   ```

## 🔌 API 使用例
Python 版のエンドポイント例は `python-app/README.md` を参照してください（Node 版はポート 3001 で同じパス）。

## 💻 ローカル開発

Python 版のローカル開発手順は `python-app/README.md` を参照してください。

### Node (Express + TypeScript + Prisma) 版

同じ PostgreSQL を共有し、Prisma はスキーマ生成のみ（マイグレーション不要）で利用します。
OTEL は `@opentelemetry/sdk-node` + auto-instrumentations で gRPC/OTLP に送信されます。

```bash
cd node-app
# 依存インストール
npm install
# Prisma クライアント生成（DB スキーマ作成はしません）
npm run prisma:generate
# 開発サーバー起動
PORT=3001 npm run dev
```

### Spring Boot 版

同じ PostgreSQL の `todos` テーブルをそのまま利用し、マイグレーションは行いません。OTEL は Java Agent で自動計装します。

```bash
cd spring-app
mvn clean package -DskipTests          # または docker-compose が自動ビルド
docker-compose build spring-api
docker-compose up -d spring-api

# エンドポイント
# GET/POST/PUT/DELETE http://localhost:8080/api/v1/todos
```

### Quarkus 版

同じ `todos` テーブルを利用する Quarkus (RESTEasy Reactive + Hibernate ORM) 版です。マイグレーションは行いません。Quarkus OTEL 拡張で OTLP にトレース/メトリクス/ログを送ります。

```bash
cd quarkus-app
cp .env.example .env
mvn -B -DskipTests package

# ローカル実行
PORT=8081 DATABASE_JDBC_URL=jdbc:postgresql://localhost:5432/tododb \
DATABASE_USERNAME=todouser DATABASE_PASSWORD=todopass \
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317 \
java -jar target/*-runner.jar

# Docker Compose から
docker-compose build quarkus-api
docker-compose up -d quarkus-api
```

### Go (chi + sqlx + pgx) 版

同じ PostgreSQL の `todos` テーブルをそのまま利用します。Docker ビルド時に Loongsuite の `otel go build` で自動計装したバイナリを生成します（トレース/メトリクス）。ログは stdout に加えてファイルにも出力でき、compose では Collector の filelog receiver で Loki に送ります。

```bash
cd go-app
cp .env.example .env

# ビルド/テスト
go test ./...

# ローカル起動
PORT=3002 DATABASE_URL=postgresql://todouser:todopass@localhost:5432/tododb \\
  go run ./cmd/server
```

### Rust (axum + sqlx) 版

同じ PostgreSQL の `todos` テーブルをそのまま利用します。compose では Rust アプリ本体に OTEL SDK を入れず、Beyla（eBPF）で HTTP/SQL を自動計装して Collector に送ります（Linux カーネル要件あり）。

```bash
cd rust-app
export DATABASE_URL=postgresql://todouser:todopass@localhost:5432/tododb
export PORT=3003
cargo run
```

## 📁 プロジェクト構造

```
hello-otel/
├── python-app/              # Python (FastAPI) 版 API 一式
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py              # FastAPIアプリ（クリーンコード）
│   │   ├── config.py            # 設定管理
│   │   ├── database.py          # DB接続
│   │   ├── models/
│   │   │   └── todo.py          # SQLAlchemyモデル
│   │   ├── schemas/
│   │   │   └── todo.py          # Pydanticスキーマ
│   │   └── routers/
│   │       └── todos.py         # CRUDエンドポイント
│   ├── alembic/                 # DBマイグレーション
│   │   └── versions/
│   ├── alembic.ini
│   ├── Dockerfile               # uv対応
│   ├── pyproject.toml           # uv依存関係
│   └── .env.example             # 環境変数テンプレート
├── node-app/                # Node (Express + TypeScript + Prisma) 版 API
│   ├── src/                     # ルーター/エントリ
│   ├── prisma/                  # Prisma schema（マイグレーションなし）
│   ├── Dockerfile
│   ├── package.json
│   └── .env.example
├── spring-app/              # Spring Boot + Maven 版 API（OTEL Java Agent）
│   ├── src/
│   ├── pom.xml
│   ├── Dockerfile
│   └── .env.example
├── quarkus-app/             # Quarkus (RESTEasy Reactive + OTEL) 版 API
│   ├── src/
│   ├── pom.xml
│   ├── Dockerfile
│   └── .env.example
├── go-app/                  # Go (chi + sqlx + pgx) 版 API
│   ├── cmd/
│   ├── internal/
│   ├── Dockerfile
│   └── .env.example
├── rust-app/                # Rust (axum + sqlx) 版 API（Beylaで自動計装）
│   ├── src/
│   ├── Cargo.toml
│   ├── Dockerfile
│   └── README.md
├── k6/                      # k6 テストスクリプト
├── beyla/                   # Beyla(eBPF) 設定
├── llm-job/                 # OpenAI 呼び出し job（Traceloop/OpenLLMetry）
├── collector.yaml               # spanmetrics 用 OTEL Collector 設定
├── docker-compose.yml           # アプリ/観測基盤構成
├── grafana-dashboard-todo.json  # Todo API用 Grafana Dashboard (importして利用)
├── grafana/                     # Grafana provisioning
└── README.md
```

## 🐛 トラブルシューティング

### アプリケーションが起動しない

```bash
# ログ確認
docker-compose logs app

# コンテナ再起動
docker-compose restart app
```

### トレースが表示されない

```bash
# 環境変数確認
docker-compose exec app env | grep OTEL

# LGTMの状態確認
docker-compose logs lgtm

# アプリ再起動
docker-compose restart app
```

### Grafana にアクセスできない

```bash
# LGTMコンテナの状態
docker-compose ps lgtm

# LGTMログ確認
docker-compose logs lgtm

# 再起動
docker-compose restart lgtm
```

### データベース接続エラー

```bash
# PostgreSQL状態確認
docker-compose ps postgres

# データベースログ
docker-compose logs postgres

# ヘルスチェック
docker-compose exec postgres pg_isready -U todouser
```

### 完全リセット

```bash
# 全コンテナ停止・削除
docker-compose down

# ボリューム含めて削除
docker-compose down -v

# 再構築
docker-compose up -d --build
```

## 📖 詳細ドキュメント

- [`docs/architecture/FINAL_ARCHITECTURE.md`](docs/architecture/FINAL_ARCHITECTURE.md) - 最終アーキテクチャ設計
- [`docs/guides/MIMIR_GUIDE.md`](docs/guides/MIMIR_GUIDE.md) - Mimir vs Prometheus 比較
- [`docs/guides/OBSERVABILITY_GUIDE.md`](docs/guides/OBSERVABILITY_GUIDE.md) - OpenTelemetry 技術ガイド
- [`docs/guides/BEST_PRACTICES.md`](docs/guides/BEST_PRACTICES.md) - コーディング規約

## 🎯 自動取得されるテレメトリ

### トレース

- HTTP リクエスト（メソッド、パス、ステータスコード）
- SQL クエリ（クエリ文、パラメータ、実行時間）
- エラー情報（スタックトレース）

### ログ

- アプリケーションログ（標準出力）
- trace_id、span_id の自動付与
- エラーログ

### メトリクス

- `http.server.duration` - リクエスト処理時間
- `http.server.active_requests` - アクティブリクエスト数
- `db.client.connections.usage` - DB 接続プール使用状況

## 🔧 開発コマンド

```bash
# コンテナ起動
docker-compose up -d

# ログ監視
docker-compose logs -f app

# シェル接続
docker-compose exec app /bin/bash

# Node API 再起動
docker-compose restart node-api
```

## 🌟 重要なポイント

### 🎯 Grafana OTEL-LGTM の利点

1. **設定ファイル不要** - すぐ使える
2. **データソース自動設定** - 手動設定不要
3. **1 コンテナで完結** - リソース効率的
4. **開発に最適** - 本番移行も容易

## 🚀 本番環境への移行

OTEL-LGTM はローカル開発用です。本番では：

- **Grafana Cloud** - マネージドサービス推奨
- **個別デプロイ** - Tempo、Loki、Mimir を分離
- **Kubernetes** - オペレーターで自動スケーリング

**重要**: アプリケーションコードは変更不要！環境変数のみ変更。

## 📚 参考資料

- [Grafana OTEL-LGTM](https://hub.docker.com/r/grafana/otel-lgtm)
- [OpenTelemetry Python](https://opentelemetry.io/docs/instrumentation/python/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [uv Documentation](https://github.com/astral-sh/uv)
