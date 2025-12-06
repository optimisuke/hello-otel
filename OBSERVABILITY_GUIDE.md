# OpenTelemetry & LGTM スタック統合ガイド

## OpenTelemetry 概要

OpenTelemetry は、アプリケーションから観測データ（トレース、メトリクス、ログ）を収集・エクスポートするための標準化されたフレームワークです。

## 3 つの観測データタイプ

### 1. トレース（Traces）

分散システム内でのリクエストの流れを追跡

**構成要素:**

- **Trace**: 複数のスパンの集合
- **Span**: 処理の単位（開始時刻、終了時刻、属性、イベント）
- **Context Propagation**: トレースコンテキストの伝播

**ユースケース:**

- リクエストのエンドツーエンド追跡
- ボトルネック特定
- サービス間の依存関係可視化

### 2. メトリクス（Metrics）

時系列の数値データ

**種類:**

- **Counter**: 増加のみ（リクエスト数など）
- **Gauge**: 増減可能（メモリ使用量など）
- **Histogram**: 値の分布（レイテンシーなど）

**ユースケース:**

- パフォーマンス監視
- キャパシティプランニング
- アラート設定

### 3. ログ（Logs）

構造化されたイベント記録

**特徴:**

- JSON 形式
- トレースコンテキスト付与
- ログレベル（DEBUG, INFO, WARNING, ERROR）

**ユースケース:**

- デバッグ
- 監査証跡
- エラー分析

## FastAPI での OpenTelemetry 計装

### トレース計装の実装パターン

#### 1. 自動計装（Auto-instrumentation）

最も簡単な方法。FastAPI と SQLAlchemy を自動的に計装します。

```python
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.sqlalchemy import SQLAlchemyInstrumentor

# FastAPI自動計装
FastAPIInstrumentor.instrument_app(app)

# SQLAlchemy自動計装
SQLAlchemyInstrumentor().instrument(engine=engine)
```

**自動取得される情報:**

- HTTP メソッド、パス、ステータスコード
- リクエスト/レスポンスヘッダー
- SQL クエリとパラメータ
- データベース接続情報

#### 2. カスタムスパン

ビジネスロジックに特化したスパンを追加します。

```python
from opentelemetry import trace

tracer = trace.get_tracer(__name__)

async def create_todo(todo_ TodoCreate, db: AsyncSession):
    with tracer.start_as_current_span("create_todo_business_logic") as span:
        # スパンに属性を追加
        span.set_attribute("todo.title", todo_data.title)
        span.set_attribute("todo.completed", todo_data.completed)

        # ビジネスロジック
        todo = Todo(**todo_data.dict())
        db.add(todo)
        await db.commit()

        # イベントを記録
        span.add_event("Todo created successfully")

        return todo
```

**カスタムスパンのベストプラクティス:**

- 意味のある名前を付ける
- 重要な属性を追加
- エラー時はスパンステータスを設定
- ネストしたスパンで詳細を記録

### メトリクス計装の実装パターン

#### 1. カウンター

```python
from opentelemetry import metrics
from opentelemetry.metrics import Counter

meter = metrics.get_meter(__name__)

# カウンターの作成
request_counter = meter.create_counter(
    name="http.server.requests",
    description="Total HTTP requests",
    unit="1"
)

# 使用例
request_counter.add(1, {"method": "GET", "endpoint": "/api/v1/todos"})
```

#### 2. ヒストグラム

```python
from opentelemetry.metrics import Histogram

request_duration = meter.create_histogram(
    name="http.server.duration",
    description="HTTP request duration",
    unit="ms"
)

# 使用例
import time
start = time.time()
# ... 処理 ...
duration = (time.time() - start) * 1000
request_duration.record(duration, {"method": "POST", "endpoint": "/api/v1/todos"})
```

#### 3. ゲージ

```python
from opentelemetry.metrics import ObservableGauge

def get_active_todos_count():
    # データベースからアクティブなTodo数を取得
    return db.query(Todo).filter(Todo.completed == False).count()

active_todos_gauge = meter.create_observable_gauge(
    name="todos.active.count",
    callbacks=[lambda options: get_active_todos_count()],
    description="Number of active todos",
    unit="1"
)
```

### ログ計装の実装パターン

#### 構造化ログ設定

```python
import logging
from opentelemetry._logs import set_logger_provider
from opentelemetry.sdk._logs import LoggerProvider
from opentelemetry.sdk._logs.export import BatchLogRecordProcessor
from opentelemetry.exporter.otlp.proto.grpc._log_exporter import OTLPLogExporter

# ロガープロバイダー設定
logger_provider = LoggerProvider()
set_logger_provider(logger_provider)

# OTLPエクスポーター
otlp_exporter = OTLPLogExporter(
    endpoint="http://otel-collector:4317",
    insecure=True
)

# バッチプロセッサー
logger_provider.add_log_record_processor(
    BatchLogRecordProcessor(otlp_exporter)
)
```

#### トレースコンテキスト付きログ

```python
from opentelemetry import trace
import logging

logger = logging.getLogger(__name__)

async def some_function():
    span = trace.get_current_span()
    span_context = span.get_span_context()

    logger.info(
        "Processing todo",
        extra={
            "trace_id": format(span_context.trace_id, "032x"),
            "span_id": format(span_context.span_id, "016x"),
            "todo_id": str(todo.id)
        }
    )
```

## OpenTelemetry Collector 設定

### 基本構成

```yaml
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
      - key: service.environment
        value: development
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

  # メトリクス → Prometheus
  prometheus:
    endpoint: "0.0.0.0:8889"

  # デバッグ用
  logging:
    loglevel: debug

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [otlp/tempo, logging]

    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [prometheus, logging]

    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch, resource]
      exporters: [loki, logging]
```

### パイプライン説明

**Receiver（受信）:**

- アプリケーションからデータを受信
- OTLP gRPC/HTTP プロトコル対応

**Processor（処理）:**

- `memory_limiter`: メモリ使用量制限
- `batch`: バッチ処理で効率化
- `resource`: リソース属性の追加/変更

**Exporter（送信）:**

- トレース → Tempo
- ログ → Loki
- メトリクス → Prometheus

## LGTM スタック設定

### Tempo 設定

```yaml
# tempo.yaml
server:
  http_listen_port: 3200

distributor:
  receivers:
    otlp:
      protocols:
        grpc:
          endpoint: 0.0.0.0:4317

storage:
  trace:
    backend: local
    local:
      path: /tmp/tempo/traces

query_frontend:
  search:
    enabled: true
```

### Loki 設定

```yaml
# loki-config.yaml
auth_enabled: false

server:
  http_listen_port: 3100

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h
```

### Prometheus 設定

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "todo-api"
    static_configs:
      - targets: ["app:8000"]
    metrics_path: "/metrics"

  - job_name: "otel-collector"
    static_configs:
      - targets: ["otel-collector:8889"]
```

### Grafana データソース設定

```yaml
# datasources.yml
apiVersion: 1

datasources:
  - name: Tempo
    type: tempo
    access: proxy
    url: http://tempo:3200
    jsonData:
      httpMethod: GET
      serviceMap:
        datasourceUid: prometheus
    uid: tempo

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    jsonData:
      derivedFields:
        - datasourceUid: tempo
          matcherRegex: "trace_id=(\\w+)"
          name: TraceID
          url: "$${__value.raw}"
    uid: loki

  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    jsonData:
      exemplarTraceIdDestinations:
        - datasourceUid: tempo
          name: trace_id
    uid: prometheus
```

## トラブルシューティングワークフロー

### 1. トレースベースのデバッグ

```
問題発生 → Grafana → Tempo
  ↓
高レイテンシーのトレース検索
  ↓
スパン詳細確認
  ↓
ボトルネック特定（DBクエリ、外部API呼び出し等）
  ↓
関連ログ参照（trace_idで検索）
  ↓
根本原因特定
```

### 2. メトリクスベースのアラート

```
異常検知（エラーレート上昇）
  ↓
Prometheusでメトリクス確認
  ↓
該当時刻のトレース検索
  ↓
エラースパン分析
  ↓
Lokiでエラーログ確認
  ↓
修正・対応
```

### 3. ログベースの調査

```
エラーログ検出（Loki）
  ↓
trace_id抽出
  ↓
Tempoでトレース全体確認
  ↓
関連メトリクス確認
  ↓
原因特定
```

## パフォーマンス最適化

### サンプリング戦略

全てのトレースを記録するとオーバーヘッドが大きいため、サンプリングを実施:

```python
from opentelemetry.sdk.trace.sampling import TraceIdRatioBased

# 10%のトレースをサンプリング
sampler = TraceIdRatioBased(0.1)
```

### バッチ処理

```python
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# バッチサイズ: 512
# タイムアウト: 5秒
processor = BatchSpanProcessor(
    exporter,
    max_queue_size=2048,
    max_export_batch_size=512,
    export_timeout_millis=30000
)
```

## ベストプラクティス

### 1. スパン命名規則

- 動詞 + 名詞: `create_todo`, `fetch_todos`
- 階層的: `db.query.select`, `api.handler.create`

### 2. 属性の標準化

- セマンティック規約に従う
- `http.method`, `http.status_code`, `db.system`

### 3. エラーハンドリング

```python
from opentelemetry.trace import Status, StatusCode

try:
    # 処理
    span.set_status(Status(StatusCode.OK))
except Exception as e:
    span.set_status(Status(StatusCode.ERROR, str(e)))
    span.record_exception(e)
    raise
```

### 4. コンテキスト伝播

- ヘッダー経由でコンテキスト伝播
- マイクロサービス間の追跡

### 5. カーディナリティ管理

- 高カーディナリティの属性は避ける
- ユーザー ID などは慎重に扱う

## まとめ

OpenTelemetry と LGTM スタックの統合により:

- **統一された観測性**: 単一のインターフェースでトレース、メトリクス、ログを管理
- **標準化**: ベンダーロックインを回避
- \*\*相関分
