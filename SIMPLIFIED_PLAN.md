# シンプル構成 実装計画

## 設計方針

### ✨ 核心原則

1. **アプリコードはクリーン** - 観測性コードを混入させない
2. **自動計装のみ** - カスタムスパン・メトリクスは不要
3. **最小構成** - Grafana Alloy + Tempo + Loki + Grafana
4. **環境変数で制御** - 設定はすべて外部化

## システム構成

```mermaid
graph TB
    Client[クライアント] --> FastAPI[FastAPI<br/>自動計装のみ]
    FastAPI --> PostgreSQL[(PostgreSQL)]

    FastAPI -->|OTLP gRPC| Alloy[Grafana Alloy]

    Alloy -->|トレース| Tempo[Tempo]
    Alloy -->|ログ| Loki[Loki]

    Tempo --> Grafana[Grafana]
    Loki --> Grafana
```

## サービス一覧（5 つのみ）

| サービス | ポート     | 役割                     |
| -------- | ---------- | ------------------------ |
| app      | 8000       | FastAPI アプリケーション |
| postgres | 5432       | データベース             |
| alloy    | 4317, 4318 | テレメトリ収集           |
| tempo    | 3200       | トレース保存             |
| loki     | 3100       | ログ保存                 |
| grafana  | 3000       | 可視化                   |

## プロジェクト構造（簡略版）

```
hello-otel/
├── app/
│   ├── __init__.py
│   ├── main.py              # FastAPIアプリ + 自動計装設定
│   ├── config.py            # 環境変数管理
│   ├── database.py          # DB接続
│   ├── models/
│   │   └── todo.py          # SQLAlchemyモデル
│   ├── schemas/
│   │   └── todo.py          # Pydanticスキーマ
│   └── routers/
│       └── todos.py         # CRUDエンドポイント
├── alembic/
│   └── versions/
├── grafana/
│   ├── datasources/
│   │   └── datasources.yml
│   └── dashboards/
│       └── todo-dashboard.json
├── alloy/
│   └── config.alloy         # Alloy設定
├── docker-compose.yml
├── Dockerfile
├── requirements.txt
├── alembic.ini
├── .env.example
└── README.md
```

## 実装ステップ

### Phase 1: 基盤セットアップ

1. プロジェクト構造作成
2. requirements.txt 作成（依存関係最小化）
3. .env.example 作成
4. docker-compose.yml 作成（5 サービスのみ）

### Phase 2: アプリケーション実装

5. データベースモデル（Todo）
6. Pydantic スキーマ
7. CRUD エンドポイント
8. **main.py で自動計装設定（3 行のみ）**

### Phase 3: 観測性設定

9. Grafana Alloy 設定
10. Grafana データソース設定
11. 基本ダッシュボード作成

### Phase 4: テスト・ドキュメント

12. 動作確認
13. README 更新

## 依存関係（最小化）

```txt
# requirements.txt - 観測性関連は最小限

# FastAPI
fastapi==0.109.0
uvicorn[standard]==0.27.0

# Database
sqlalchemy==2.0.25
asyncpg==0.29.0
alembic==1.13.1
psycopg2-binary==2.9.9

# OpenTelemetry - 自動計装のみ
opentelemetry-distro==0.43b0
opentelemetry-instrumentation-fastapi==0.43b0
opentelemetry-instrumentation-sqlalchemy==0.43b0
opentelemetry-exporter-otlp==1.22.0

# Validation
pydantic==2.5.3
pydantic-settings==2.1.0
python-dotenv==1.0.0
```

## main.py（観測性設定はこれだけ）

```python
from fastapi import FastAPI
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor
from app.routers import todos
from app.database import engine

# FastAPIアプリ
app = FastAPI(
    title="Todo API",
    description="Simple Todo API with auto-instrumentation"
)

# ルーター登録
app.include_router(todos.router, prefix="/api/v1/todos", tags=["todos"])

# 自動計装（これだけ！）
FastAPIInstrumentor.instrument_app(app)
SQLAlchemyInstrumentor().instrument(engine=engine)

@app.get("/health")
async def health_check():
    return {"status": "healthy"}
```

## 環境変数設定

```bash
# .env
# Database
DATABASE_URL=postgresql+asyncpg://todouser:todopass@postgres:5432/tododb

# OpenTelemetry（自動計装用）
OTEL_SERVICE_NAME=todo-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://alloy:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_TRACES_EXPORTER=otlp
OTEL_LOGS_EXPORTER=otlp
OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED=true

# Application
APP_HOST=0.0.0.0
APP_PORT=8000
LOG_LEVEL=INFO
```

## Grafana Alloy 設定（簡略版）

```hcl
// config.alloy
otelcol.receiver.otlp "default" {
  grpc {
    endpoint = "0.0.0.0:4317"
  }
  http {
    endpoint = "0.0.0.0:4318"
  }

  output {
    traces  = [otelcol.exporter.otlp.tempo.input]
    logs    = [otelcol.exporter.loki.default.input]
  }
}

// トレース → Tempo
otelcol.exporter.otlp "tempo" {
  client {
    endpoint = "tempo:4317"
    tls {
      insecure = true
    }
  }
}

// ログ → Loki
otelcol.exporter.loki "default" {
  forward_to = [loki.write.default.receiver]
}

loki.write "default" {
  endpoint {
    url = "http://loki:3100/loki/api/v1/push"
  }
}
```

## Docker Compose（簡略版）

```yaml
version: '3.8'

services:
  # FastAPI アプリケーション
  app:
    build: .
    ports:
      - "8000:8000"
    environment:
      - DATABASE_URL=postgresql+asyncpg://todouser:todopass@postgres:5432/tododb
      - OTEL_SERVICE_NAME=todo-api
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://alloy:4317
    depends_on:
      - postgres
      - alloy

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
      - postgres_/var/lib/postgresql/data

  # Grafana Alloy（テレメトリ収集）
  alloy:
    image: grafana/alloy:latest
    ports:
      - "4317:4317"  # OTLP gRPC
      - "4318:4318"  # OTLP HTTP
    volumes:
      - ./alloy/config.alloy:/etc/alloy/config.alloy
    command: run --server.http.listen-addr=0.0.0.0:12345 /etc/alloy/config.alloy

  # Tempo（トレース）
  tempo:
    image: grafana/tempo:2.3.1
    ports:
      - "3200:3200"
      - "4317"
    command: ["-config.file=/etc/tempo.yaml"]
    volumes:
      - ./tempo/tempo.yaml:/etc/tempo.yaml
      - tempo_/tmp/tempo

  # Loki（ログ）
  loki:
    image: grafana/loki:2.9.3
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki_/loki

  # Grafana（可視化）
  grafana:
    image: grafana/grafana:10.2.3
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    volumes:
      - ./grafana/datasources:/etc/grafana/provisioning/datasources
      - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
      - grafana_/var/lib/grafana

volumes:
  postgres_
  tempo_data:
  loki_
  grafana_
```

## 実装時の注意点

### ✅ やること

- 環境変数で OpenTelemetry 設定
- FastAPI/SQLAlchemy の自動計装のみ使用
- ビジネスロジックに集中

### ❌ やらないこと

- カスタムスパン作成
- 手動でのメトリクス記録
- ログに特別な設定を追加
- トレースコンテキストの手動伝播

## 自動的に取得されるデータ

### トレース

- HTTP リクエスト（メソッド、パス、ステータス）
- SQL クエリ（クエリ文、パラメータ、実行時間）
- エラー情報（スタックトレース）

### ログ

- アプリケーションログ（標準出力）
- エラーログ
- トレース ID/スパン ID の自動付与

## 動作確認手順

1. **起動**

```bash
docker-compose up -d
docker-compose exec app alembic upgrade head
```

2. **API テスト**

```bash
# Todo作成
curl -X POST http://localhost:8000/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Todo"}'

# Todo一覧
curl http://localhost:8000/api/v1/todos
```

3. **Grafana 確認**

- http://localhost:3000
- Explore → Tempo でトレース確認
- Explore → Loki でログ確認

## 拡張性

将来的に必要になったら追加可能：

- カスタムスパン（特定の処理を詳細にトレース）
- メトリクス（Prometheus 追加）
- より詳細なダッシュボード
- アラート設定

## まとめ

この構成の利点：

1. **シンプル** - 5 サービスのみ
2. **クリーンコード** - ビジネスロジックと観測性の分離
3. **自動化** - 手動設定不要
4. **十分な機能** - トレース + ログで大半のユースケースをカバー
5. **拡張可能** - 必要に応じて後から機能追加可能

アプリケーション開発者は**通常の FastAPI コードを書くだけ**で、完全な観測性が得られます。
