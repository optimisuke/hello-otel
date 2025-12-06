# 実装計画詳細

## フェーズ 1: プロジェクト基盤構築

### 1.1 プロジェクト構造の作成

必要なディレクトリとファイル:

```
hello-otel/
├── app/
│   ├── __init__.py
│   ├── main.py
│   ├── config.py
│   ├── database.py
│   ├── models/
│   ├── schemas/
│   ├── routers/
│   ├── services/
│   └── observability/
├── alembic/
├── grafana/
├── otel-collector/
├── docker-compose.yml
├── Dockerfile
├── requirements.txt
├── alembic.ini
├── .env.example
└── README.md
```

### 1.2 依存関係（requirements.txt）

```
# FastAPI & ASGI Server
fastapi==0.109.0
uvicorn[standard]==0.27.0

# Database
sqlalchemy==2.0.25
asyncpg==0.29.0
alembic==1.13.1
psycopg2-binary==2.9.9

# OpenTelemetry Core
opentelemetry-api==1.22.0
opentelemetry-sdk==1.22.0
opentelemetry-instrumentation==0.43b0

# OpenTelemetry Instrumentations
opentelemetry-instrumentation-fastapi==0.43b0
opentelemetry-instrumentation-sqlalchemy==0.43b0
opentelemetry-instrumentation-logging==0.43b0
opentelemetry-instrumentation-system-metrics==0.43b0

# OpenTelemetry Exporters
opentelemetry-exporter-otlp-proto-grpc==1.22.0
opentelemetry-exporter-prometheus==0.43b0

# Validation & Utilities
pydantic==2.5.3
pydantic-settings==2.1.0
python-dotenv==1.0.0

# Monitoring
prometheus-client==0.19.0
```

## フェーズ 2: データベース設定

### 2.1 SQLAlchemy モデル（app/models/todo.py）

```python
from sqlalchemy import Column, String, Boolean, DateTime, Text
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.sql import func
import uuid
from app.database import Base

class Todo(Base):
    __tablename__ = "todos"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    title = Column(String(200), nullable=False, index=True)
    description = Column(Text, nullable=True)
    completed = Column(Boolean, default=False, nullable=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now())
```

### 2.2 Pydantic スキーマ（app/schemas/todo.py）

```python
from pydantic import BaseModel, Field
from datetime import datetime
from uuid import UUID
from typing import Optional

class TodoBase(BaseModel):
    title: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = None
    completed: bool = False

class TodoCreate(TodoBase):
    pass

class TodoUpdate(BaseModel):
    title: Optional[str] = Field(None, min_length=1, max_length=200)
    description: Optional[str] = None
    completed: Optional[bool] = None

class TodoResponse(TodoBase):
    id: UUID
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True
```

## フェーズ 3: OpenTelemetry 計装

### 3.1 トレース設定（app/observability/tracing.py）

主要機能:

- FastAPI 自動計装
- SQLAlchemy 自動計装
- カスタムスパン作成
- OTLP gRPC エクスポーター設定

### 3.2 メトリクス設定（app/observability/metrics.py）

収集メトリクス:

- `http_requests_total`: HTTP リクエスト総数
- `http_request_duration_seconds`: リクエストレイテンシー
- `todos_total`: Todo 項目総数
- `todos_completed_ratio`: 完了率
- `db_query_duration_seconds`: データベースクエリ時間

### 3.3 ログ設定（app/observability/logging.py）

機能:

- 構造化 JSON 形式ログ
- trace_id、span_id の自動付与
- OTLP 経由で Loki へ送信
- ログレベル別フィルタリング

## フェーズ 4: Docker Compose 構成

### 4.1 サービス構成

#### PostgreSQL

```yaml
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
```

#### OpenTelemetry Collector

```yaml
otel-collector:
  image: otel/opentelemetry-collector-contrib:0.91.0
  command: ["--config=/etc/otel-collector-config.yaml"]
  volumes:
    - ./otel-collector/config.yaml:/etc/otel-collector-config.yaml
  ports:
    - "4317:4317" # OTLP gRPC
    - "4318:4318" # OTLP HTTP
```

#### Tempo

```yaml
tempo:
  image: grafana/tempo:2.3.1
  command: ["-config.file=/etc/tempo.yaml"]
  volumes:
    - ./tempo/tempo.yaml:/etc/tempo.yaml
    - tempo_/tmp/tempo
  ports:
    - "3200:3200" # Tempo
    - "4317" # OTLP gRPC
```

#### Loki

```yaml
loki:
  image: grafana/loki:2.9.3
  ports:
    - "3100:3100"
  command: -config.file=/etc/loki/local-config.yaml
  volumes:
    - loki_/loki
```

#### Prometheus

```yaml
prometheus:
  image: prom/prometheus:v2.48.1
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus_/prometheus
  ports:
    - "9090:9090"
```

#### Grafana

```yaml
grafana:
  image: grafana/grafana:10.2.3
  ports:
    - "3000:3000"
  environment:
    - GF_SECURITY_ADMIN_PASSWORD=admin
  volumes:
    - ./grafana/datasources:/etc/grafana/provisioning/datasources
    - ./grafana/dashboards:/etc/grafana/provisioning/dashboards
    - grafana_/var/lib/grafana
```

### 4.2 OTel Collector 設定

レシーバー:

- OTLP (gRPC/HTTP)

プロセッサー:

- Batch
- Memory Limiter
- Resource Detection

エクスポーター:

- Tempo (トレース)
- Loki (ログ)
- Prometheus (メトリクス)

## フェーズ 5: API 実装

### 5.1 エンドポイント実装（app/routers/todos.py）

```python
@router.get("/", response_model=List[TodoResponse])
async def get_todos(
    skip: int = 0,
    limit: int = 100,
    db: AsyncSession = Depends(get_db)
) -> List[Todo]:
    # トレーシング
    # ビジネスロジック
    # エラーハンドリング

@router.get("/{todo_id}", response_model=TodoResponse)
async def get_todo(todo_id: UUID, db: AsyncSession = Depends(get_db)):
    # 実装

@router.post("/", response_model=TodoResponse, status_code=201)
async def create_todo(todo: TodoCreate, db: AsyncSession = Depends(get_db)):
    # 実装

@router.put("/{todo_id}", response_model=TodoResponse)
async def update_todo(todo_id: UUID, todo: TodoUpdate, db: AsyncSession = Depends(get_db)):
    # 実装

@router.delete("/{todo_id}", status_code=204)
async def delete_todo(todo_id: UUID, db: AsyncSession = Depends(get_db)):
    # 実装
```

### 5.2 サービス層（app/services/todo_service.py）

ビジネスロジックの分離:

- CRUD 操作
- バリデーション
- エラーハンドリング
- カスタムスパン作成

## フェーズ 6: Grafana ダッシュボード

### 6.1 データソース設定

```yaml
apiVersion: 1
datasources:
  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100

  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
```

### 6.2 ダッシュボードパネル構成

1. **概要統計**

   - Total Requests (24h)
   - Avg Response Time
   - Error Rate
   - Total Todos / Completed

2. **トレース分析**

   - Request Rate by Endpoint
   - Latency Heatmap
   - Trace Timeline

3. **メトリクス**

   - HTTP Request Rate
   - Response Time Percentiles (p50, p95, p99)
   - Database Query Duration

4. **ログ**
   - Error Logs
   - Recent Activity
   - Trace ID Filter

## フェーズ 7: テスト・検証

### 7.1 動作確認手順

1. Docker Compose 起動
2. データベースマイグレーション
3. API エンドポイントテスト（curl/Postman）
4. トレースの可視化確認（Grafana）
5. メトリクスの収集確認
6. ログの集約確認

### 7.2 サンプルリクエスト

```bash
# Todo作成
curl -X POST http://localhost:8000/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Todo", "description": "Test Description"}'

# Todo一覧取得
curl http://localhost:8000/api/v1/todos

# Todo更新
curl -X PUT http://localhost:8000/api/v1/todos/{id} \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

## フェーズ 8: ドキュメント整備

### 8.1 README.md

内容:

- プロジェクト概要
- クイックスタート
- 前提条件
- セットアップ手順
- API 使用例
- Grafana アクセス方法
- トラブルシューティング

### 8.2 環境変数（.env.example）

```
# Database
DATABASE_URL=postgresql+asyncpg://todouser:todopass@postgres:5432/tododb

# OpenTelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_SERVICE_NAME=todo-api
OTEL_TRACES_EXPORTER=otlp
OTEL_METRICS_EXPORTER=otlp
OTEL_LOGS_EXPORTER=otlp

# Application
APP_HOST=0.0.0.0
APP_PORT=8000
LOG_LEVEL=INFO
```

## 実装順序まとめ

1. ✅ プロジェクト構造作成
2. ✅ 依存関係定義（requirements.txt）
3. ✅ 環境変数設定（.env.example）
4. ✅ データベースモデル・スキーマ
5. ✅ OpenTelemetry 計装設定
6. ✅ FastAPI アプリケーション基本構造
7. ✅ Todo CRUD 実装
8. ✅ Docker Compose 設定
9. ✅ OTel Collector 設定
10. ✅ Grafana データソース・ダッシュボード
11. ✅ ドキュメント作成
12. ✅ テスト・検証

## 技術的考慮事項

### パフォーマンス

- 非同期 I/O（asyncio）
- 接続プーリング
- バッチ処理（OTel）

### セキュリティ

- 環境変数による機密情報管理
- SQL インジェクション対策
- CORS 設定

### 運用

- ヘルスチェックエンドポイント
- Graceful Shutdown
- ログローテーション

### 拡張性

- マイクロサービス対応可能な設計
- 水平スケーリング可能
- プラグイン可能なアーキテクチャ
