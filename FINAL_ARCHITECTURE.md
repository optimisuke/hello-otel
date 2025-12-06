# 最終アーキテクチャ（Grafana OTEL-LGTM 採用）

## 🎯 設計方針の変更

**重要な発見**: Grafana 公式が提供する **`grafana/otel-lgtm`** イメージを採用することで、構成を劇的にシンプル化！

### Before（当初の設計）

```
7サービス: app, postgres, otel-collector, tempo, loki, prometheus, grafana
```

### After（最終設計）✅

```
3サービス: app, postgres, lgtm (all-in-one)
```

## システム構成

```mermaid
graph TB
    Client[クライアント] --> FastAPI[FastAPI アプリ<br/>コード変更ゼロ]
    FastAPI --> PostgreSQL[(PostgreSQL)]

    FastAPI -->|OTLP gRPC| LGTM[Grafana OTEL-LGTM<br/>統合イメージ]

    subgraph "LGTM (1コンテナ)"
        OTel[OpenTelemetry<br/>Collector]
        Tempo[Tempo]
        Loki[Loki]
        Mimir[Mimir]
        Grafana[Grafana]

        OTel --> Tempo
        OTel --> Loki
        OTel --> Mimir

        Tempo --> Grafana
        Loki --> Grafana
        Mimir --> Grafana
    end

    User[ユーザー] --> Grafana
```

## サービス一覧（3 つのみ！）

| サービス | ポート           | 役割             | イメージ              |
| -------- | ---------------- | ---------------- | --------------------- |
| app      | 8000             | FastAPI アプリ   | python:3.11-slim + uv |
| postgres | 5432             | データベース     | postgres:16-alpine    |
| **lgtm** | 3000, 4317, 4318 | **統合観測基盤** | **grafana/otel-lgtm** |

## Grafana OTEL-LGTM の特徴

### 含まれるコンポーネント

- ✅ **OpenTelemetry Collector** - テレメトリ収集
- ✅ **Tempo** - トレース保存
- ✅ **Loki** - ログ保存
- ✅ **Mimir** - メトリクス保存
- ✅ **Grafana** - 統合 UI

### メリット

1. **極めてシンプル** - 設定ファイル不要
2. **自動統合** - データソースが事前設定済み
3. **開発に最適** - すぐに使える
4. **リソース効率** - 1 コンテナで完結

### 公式情報

- Docker Hub: https://hub.docker.com/r/grafana/otel-lgtm
- 用途: ローカル開発・デモ・学習

## プロジェクト構造（簡略化）

```
hello-otel/
├── app/
│   ├── __init__.py
│   ├── main.py              # クリーンなFastAPIコード
│   ├── config.py
│   ├── database.py
│   ├── models/
│   │   └── todo.py
│   ├── schemas/
│   │   └── todo.py
│   └── routers/
│       └── todos.py
├── alembic/
│   └── versions/
├── docker-compose.yml       # 3サービスのみ
├── Dockerfile
├── pyproject.toml
├── .env.example
└── README.md
```

## Docker Compose 設定（最終版）

```yaml
version: "3.8"

services:
  app:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql+asyncpg://todouser:todopass@postgres:5432/tododb
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://lgtm:4317
      - OTEL_SERVICE_NAME=todo-api
    depends_on:
      - postgres
      - lgtm

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: todouser
      POSTGRES_PASSWORD: todopass
      POSTGRES_DB: tododb
    ports:
      - "5432:5432"
    volumes:
      - postgres_/var/lib/postgresql/data

  lgtm:
    image: grafana/otel-lgtm:latest
    ports:
      - "3000:3000" # Grafana UI
      - "4317:4317" # OTLP gRPC
      - "4318:4318" # OTLP HTTP
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

**これだけ！** 設定ファイル不要！

## 環境変数（.env）

```bash
# Database
DATABASE_URL=postgresql+asyncpg://todouser:todopass@postgres:5432/tododb

# OpenTelemetry（自動計装用）
OTEL_EXPORTER_OTLP_ENDPOINT=http://lgtm:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_SERVICE_NAME=todo-api
OTEL_RESOURCE_ATTRIBUTES=deployment.environment=development

# OpenTelemetry Exporters
OTEL_TRACES_EXPORTER=otlp
OTEL_METRICS_EXPORTER=otlp
OTEL_LOGS_EXPORTER=otlp

# Python Auto-Instrumentation
OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED=true
```

## Dockerfile（uv 対応）

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

# Install dependencies
COPY pyproject.toml ./
RUN uv pip install --system -r pyproject.toml

# Copy application
COPY . .

EXPOSE 8000

# Run with OpenTelemetry auto-instrumentation
CMD ["opentelemetry-instrument", \
     "--traces_exporter", "otlp", \
     "--metrics_exporter", "otlp", \
     "--logs_exporter", "otlp", \
     "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

## main.py（完全にクリーン）

```python
"""
Todo API with Zero-Code Observability
opentelemetry-instrumentコマンドが全自動で計装
"""
from fastapi import FastAPI
from app.routers import todos

app = FastAPI(
    title="Todo API",
    description="Simple Todo API with Automatic Observability",
    version="0.1.0"
)

# ルーター登録
app.include_router(
    todos.router,
    prefix="/api/v1/todos",
    tags=["todos"]
)

@app.get("/")
async def root():
    return {"message": "Todo API with OTEL-LGTM"}

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

# 観測性コードは一切なし！
```

## クイックスタート

### 1. 起動

```bash
# コンテナ起動
docker-compose up -d

# データベースマイグレーション
docker-compose exec app alembic upgrade head
```

### 2. アクセス

| サービス    | URL                        | 用途                   |
| ----------- | -------------------------- | ---------------------- |
| API         | http://localhost:8000      | FastAPI                |
| API Docs    | http://localhost:8000/docs | Swagger UI             |
| **Grafana** | **http://localhost:3000**  | **統合ダッシュボード** |

**Grafana ログイン**

- ユーザー: `admin`
- パスワード: `admin`

### 3. 観測性の確認

Grafana（http://localhost:3000）にアクセス：

1. **Explore** → **Tempo** → トレース検索
2. **Explore** → **Loki** → ログ検索
3. **Explore** → **Mimir** → メトリクス確認

すべてのデータソースが自動設定済み！

## 自動取得されるテレメトリ

### トレース

- HTTP リクエスト（メソッド、パス、ステータス）
- SQL クエリ（クエリ文、実行時間）
- エラー（スタックトレース）

### ログ

- アプリケーションログ
- trace_id/span_id 自動付与
- エラーログ

### メトリクス

- `http.server.duration` - リクエストレイテンシー
- `http.server.active_requests` - アクティブリクエスト
- `db.client.connections.usage` - DB 接続

## 開発コマンド

```bash
# ローカル開発
uv run uvicorn app.main:app --reload

# OpenTelemetry自動計装で実行
uv run opentelemetry-instrument uvicorn app.main:app --host 0.0.0.0 --port 8000

# マイグレーション作成
docker-compose exec app alembic revision --autogenerate -m "description"

# マイグレーション実行
docker-compose exec app alembic upgrade head

# ログ確認
docker-compose logs -f app
docker-compose logs -f lgtm
```

## トラブルシューティング

### トレースが表示されない

```bash
# アプリのログ確認
docker-compose logs app

# LGTMの状態確認
docker-compose logs lgtm

# 環境変数確認
docker-compose exec app env | grep OTEL
```

### Grafana にアクセスできない

```bash
# LGTMコンテナの状態確認
docker-compose ps lgtm

# 再起動
docker-compose restart lgtm
```

## 実装時の注意点

### ✅ やること

- 通常の FastAPI コードを書く
- 環境変数で OpenTelemetry 設定
- `opentelemetry-instrument`コマンドで起動

### ❌ やらないこと

- OpenTelemetry の import
- 手動でのスパン作成（自動計装で十分）
- 設定ファイルの作成（LGTM が全自動）
- データソースの手動設定（事前設定済み）

## 本番環境への移行

OTEL-LGTM はローカル開発用です。本番環境では：

1. **Grafana Cloud** - マネージドサービス
2. **個別デプロイ** - Tempo、Loki、Mimir を分離
3. **Kubernetes オペレーター** - 自動スケーリング

本番移行時もアプリケーションコードは変更不要！

## まとめ

### 🎉 達成したこと

1. **最もシンプル** - わずか 3 サービス
2. **設定ファイル不要** - docker-compose.yml のみ
3. **完全な観測性** - トレース + ログ + メトリクス
4. **クリーンコード** - アプリに観測性コードゼロ
5. **すぐに使える** - 起動後すぐ Grafana で確認可能

### 📚 参考資料

- [Grafana OTEL-LGTM](https://hub.docker.com/r/grafana/otel-lgtm)
- [OpenTelemetry Python](https://opentelemetry.io/docs/instrumentation/python/)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)

---

**このアーキテクチャで、最もシンプルかつ完全な観測性を実現！**
