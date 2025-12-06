# 最終アーキテクチャ（最もシンプル＆標準的）

## 設計原則

1. ✅ **アプリコードは完全にクリーン** - import 不要、コード変更ゼロ
2. ✅ **`opentelemetry-instrument`コマンド** - 完全自動計装
3. ✅ **uv で Python 管理** - pyproject.toml + uv.lock
4. ✅ **OpenTelemetry Collector** - 標準的で拡張性が高い
5. ✅ **完全な LGTM Stack** - Loki + Grafana + Tempo + Mimir

## システム構成（6 サービス）

```mermaid
graph TB
    Client[クライアント] --> FastAPI[FastAPI アプリ<br/>コード変更なし]
    FastAPI --> PostgreSQL[(PostgreSQL)]

    FastAPI -->|OTLP gRPC| Collector[OpenTelemetry<br/>Collector]

    Collector -->|トレース| Tempo[Tempo]
    Collector -->|ログ| Loki[Loki]
    Collector -->|メトリクス| Mimir[Mimir]

    Tempo --> Grafana[Grafana]
    Loki --> Grafana
    Mimir --> Grafana

    User[ユーザー] --> Grafana
```

## サービス一覧

| サービス       | ポート     | 役割           | イメージ                             |
| -------------- | ---------- | -------------- | ------------------------------------ |
| app            | 8000       | FastAPI アプリ | python:3.11-slim                     |
| postgres       | 5432       | データベース   | postgres:16-alpine                   |
| otel-collector | 4317, 4318 | テレメトリ収集 | otel/opentelemetry-collector-contrib |
| tempo          | 3200       | トレース保存   | grafana/tempo                        |
| loki           | 3100       | ログ保存       | grafana/loki                         |
| mimir          | 9009       | メトリクス保存 | grafana/mimir                        |
| grafana        | 3000       | 可視化         | grafana/grafana                      |

## プロジェクト構造（最小限）

```
hello-otel/
├── app/
│   ├── __init__.py
│   ├── main.py              # 純粋なFastAPIコード（import不要）
│   ├── config.py            # 環境変数
│   ├── database.py          # DB接続
│   ├── models/
│   │   └── todo.py
│   ├── schemas/
│   │   └── todo.py
│   └── routers/
│       └── todos.py
├── alembic/
│   └── versions/
├── grafana/
│   ├── datasources/
│   │   └── datasources.yml
│   └── dashboards/
│       └── lgtm-dashboard.json
├── otel-collector/
│   └── config.yaml          # Collector設定
├── docker-compose.yml
├── Dockerfile
├── pyproject.toml           # uv管理
├── uv.lock
├── alembic.ini
├── .env.example
└── README.md
```

## Python 依存関係管理（uv）

### pyproject.toml

```toml
[project]
name = "hello-otel"
version = "0.1.0"
description = "Todo API with OpenTelemetry and LGTM Stack"
requires-python = ">=3.11"

dependencies = [
    # FastAPI
    "fastapi>=0.109.0",
    "uvicorn[standard]>=0.27.0",

    # Database
    "sqlalchemy>=2.0.25",
    "asyncpg>=0.29.0",
    "alembic>=1.13.1",
    "psycopg2-binary>=2.9.9",

    # OpenTelemetry - 自動計装パッケージ
    "opentelemetry-distro>=0.43b0",
    "opentelemetry-instrumentation>=0.43b0",
    "opentelemetry-exporter-otlp>=1.22.0",

    # Validation
    "pydantic>=2.5.3",
    "pydantic-settings>=2.1.0",
    "python-dotenv>=1.0.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.4.0",
    "pytest-asyncio>=0.21.0",
    "httpx>=0.25.0",
]

[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[tool.uv]
dev-dependencies = [
    "pytest>=7.4.0",
    "pytest-asyncio>=0.21.0",
]
```

### uv コマンド

```bash
# プロジェクト初期化
uv init

# 依存関係インストール
uv sync

# 実行
uv run uvicorn app.main:app --reload

# OpenTelemetry自動計装で実行
uv run opentelemetry-instrument uvicorn app.main:app --host 0.0.0.0 --port 8000
```

## main.py（完全にクリーン）

```python
"""
Todo API - OpenTelemetry自動計装版
コードに観測性の設定は一切なし！
"""
from fastapi import FastAPI
from app.routers import todos
from app.database import engine

# 純粋なFastAPIアプリケーション
app = FastAPI(
    title="Todo API",
    description="Simple Todo API with Zero-Code Observability",
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
    return {"message": "Todo API with OpenTelemetry"}

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

# 観測性の設定は一切なし！
# opentelemetry-instrumentコマンドが全て自動で行う
```

## Dockerfile（uv 対応）

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# uvのインストール
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

# 依存関係ファイルコピー
COPY pyproject.toml uv.lock ./

# 依存関係インストール
RUN uv sync --frozen --no-dev

# アプリケーションコピー
COPY . .

# OpenTelemetry自動計装で起動
CMD ["uv", "run", "opentelemetry-instrument", \
     "--traces_exporter", "otlp", \
     "--metrics_exporter", "otlp", \
     "--logs_exporter", "otlp", \
     "--service_name", "todo-api", \
     "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

## 環境変数（.env）

```bash
# Database
DATABASE_URL=postgresql+asyncpg://todouser:todopass@postgres:5432/tododb

# OpenTelemetry（自動計装用）
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_SERVICE_NAME=todo-api
OTEL_RESOURCE_ATTRIBUTES=deployment.environment=development

# 自動計装の詳細設定
OTEL_TRACES_EXPORTER=otlp
OTEL_METRICS_EXPORTER=otlp
OTEL_LOGS_EXPORTER=otlp
OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED=true

# FastAPIの自動計装設定
OTEL_PYTHON_FASTAPI_EXCLUDED_URLS=/health,/metrics

# Application
APP_HOST=0.0.0.0
APP_PORT=8000
LOG_LEVEL=INFO
```

## OpenTelemetry Collector 設定

```yaml
# otel-collector/config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    timeout: 10s
    send_batch_size: 1024

  memory_limiter:
    check_interval: 1s
    limit_mib: 512

  resource:
    attributes:
      - key: service.namespace
        value: todo-app
        action: upsert

exporters:
  # トレース → Tempo
  otlp/tempo:
    endpoint: tempo:4317
    tls:
      insecure: true

  # ログ → Loki
  loki:
    endpoint: http://loki:3100/loki/api/v1/push
    labels:
      resource:
        service.name: "service_name"
      attributes:
        level: "level"

  # メトリクス → Mimir
  otlphttp/mimir:
    endpoint: http://mimir:9009/otlp
    tls:
      insecure: true

  # デバッグ用
  logging:
    loglevel: info

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [otlp/tempo, logging]

    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [otlphttp/mimir, logging]

    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [loki, logging]
```

## Docker Compose（完全版）

```yaml
version: "3.8"

services:
  # FastAPI アプリケーション
  app:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql+asyncpg://todouser:todopass@postgres:5432/tododb
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
      - OTEL_SERVICE_NAME=todo-api
    depends_on:
      - postgres
      - otel-collector
    networks:
      - lgtm

  # PostgreSQL
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: todouser
      POSTGRES_PASSWORD: todopass
      POSTGRES_DB: tododb
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - lgtm

  # OpenTelemetry Collector
  otel-collector:
    image: otel/opentelemetry-collector-contrib:0.91.0
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector/config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317" # OTLP gRPC
      - "4318:4318" # OTLP HTTP
    networks:
      - lgtm

  # Tempo（トレース）
  tempo:
    image: grafana/tempo:2.3.1
    command: ["-config.file=/etc/tempo.yaml"]
    volumes:
      - ./tempo/tempo.yaml:/etc/tempo.yaml
      - tempo_/tmp/tempo
    ports:
      - "3200:3200" # Tempo
      - "4317" # OTLP gRPC
    networks:
      - lgtm

  # Loki（ログ）
  loki:
    image: grafana/loki:2.9.3
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki_/loki
    networks:
      - lgtm

  # Mimir（メトリクス）
  mimir:
    image: grafana/mimir:2.10.4
    command:
      - -config.file=/etc/mimir.yaml
    volumes:
      - ./mimir/mimir.yaml:/etc/mimir.yaml
      - mimir_/data
    ports:
      - "9009:9009"
    networks:
      - lgtm

  # Grafana（可視化）
  grafana:
    image: grafana/grafana:10.2.3
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
      - G
```
