# Grafana Mimir 完全ガイド

## Mimir とは？

Grafana Mimir は、Prometheus の後継として設計された**大規模分散メトリクスシステム**です。

### 特徴

- **水平スケーラブル** - 数百万のメトリクスシリーズを扱える
- **高可用性** - マルチテナント対応
- **Prometheus 互換** - PromQL をそのまま使用可能
- **長期保存** - オブジェクトストレージ（S3 等）に保存可能

### Prometheus と Mimir の比較

| 項目         | Prometheus       | Mimir                    |
| ------------ | ---------------- | ------------------------ |
| スケール     | 単一ノード       | 分散システム             |
| ストレージ   | ローカルディスク | オブジェクトストレージ   |
| ユースケース | 小〜中規模       | 大規模・エンタープライズ |
| 設定複雑度   | シンプル         | 複雑                     |

## ローカル開発での課題と選択肢

### 🤔 実は...ローカル開発にはオーバースペック

**結論: ローカル開発環境では Prometheus の方がシンプルで適切です**

理由：

1. Mimir は複数のマイクロサービス構成（9 個以上）
2. 設定が複雑
3. リソース消費が大きい
4. ローカル開発には不要な機能が多い

## 推奨: 3 つの選択肢

### 選択肢 1: Prometheus（推奨 - シンプル）

**メリット:**

- 設定が簡単
- 1 つのコンテナのみ
- リソース消費が少ない
- ローカル開発に最適

**デメリット:**

- 大規模本番環境には向かない

```yaml
# docker-compose.yml
prometheus:
  image: prom/prometheus:v2.48.1
  ports:
    - "9090:9090"
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
    - prometheus_/prometheus
  command:
    - "--config.file=/etc/prometheus/prometheus.yml"
    - "--storage.tsdb.path=/prometheus"
```

```yaml
# prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: "otel-collector"
    static_configs:
      - targets: ["otel-collector:8888"] # Collector自身のメトリクス

  - job_name: "todo-api"
    static_configs:
      - targets: ["app:8000"] # アプリのメトリクスエンドポイント（オプション）
```

### 選択肢 2: Mimir Monolithic Mode（中間）

**単一コンテナで動作する Mimir**

```yaml
# docker-compose.yml
mimir:
  image: grafana/mimir:2.10.4
  command:
    - -config.file=/etc/mimir/mimir.yaml
    - -target=all
  ports:
    - "9009:9009"
  volumes:
    - ./mimir/mimir.yaml:/etc/mimir/mimir.yaml
    - mimir_/data
```

```yaml
# mimir/mimir.yaml
target: all

server:
  http_listen_port: 9009
  grpc_listen_port: 9095

common:
  storage:
    backend: filesystem
    filesystem:
      dir: /data/mimir

blocks_storage:
  backend: filesystem
  filesystem:
    dir: /data/blocks

compactor:
  data_dir: /data/compactor

distributor:
  ring:
    kvstore:
      store: memberlist

ingester:
  ring:
    kvstore:
      store: memberlist

ruler_storage:
  backend: filesystem
  filesystem:
    dir: /data/ruler

memberlist:
  join_members:
    - mimir:7946

limits:
  ingestion_rate: 100000
  ingestion_burst_size: 200000
```

### 選択肢 3: メトリクス不要（最もシンプル）

トレースとログだけで十分な場合：

- Tempo - トレース
- Loki - ログ
- Grafana - 可視化

メトリクスは後から追加可能！

## OpenTelemetry Collector 設定

### Prometheus の場合

```yaml
exporters:
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: todo_api

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheus]
```

Prometheus の設定：

```yaml
scrape_configs:
  - job_name: "otel-collector"
    static_configs:
      - targets: ["otel-collector:8889"]
```

### Mimir の場合

```yaml
exporters:
  otlphttp/mimir:
    endpoint: http://mimir:9009/otlp
    tls:
      insecure: true

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp/mimir]
```

## 自動計装で取得できるメトリクス

`opentelemetry-instrument` コマンドで自動的に取得されるメトリクス：

### HTTP メトリクス

- `http.server.active_requests` - アクティブなリクエスト数
- `http.server.duration` - リクエスト処理時間
- `http.server.request.size` - リクエストサイズ
- `http.server.response.size` - レスポンスサイズ

### システムメトリクス

- `process.runtime.cpython.memory` - メモリ使用量
- `process.runtime.cpython.cpu_time` - CPU 時間
- `process.runtime.cpython.gc_count` - ガベージコレクション回数

### データベースメトリクス（SQLAlchemy）

- `db.client.connections.usage` - コネクション使用状況
- `db.client.connections.idle` - アイドルコネクション数

## Grafana での確認方法

### データソース設定

#### Prometheus の場合

```yaml
# grafana/datasources/datasources.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: false
```

#### Mimir の場合

```yaml
# grafana/datasources/datasources.yml
apiVersion: 1

datasources:
  - name: Mimir
    type: prometheus # Prometheus互換
    access: proxy
    url: http://mimir:9009/prometheus
    jsonData:
      httpMethod: POST
    isDefault: false
```

### クエリ例

```promql
# リクエストレート（1分あたり）
rate(http_server_duration_count[1m])

# 平均レスポンスタイム
rate(http_server_duration_sum[5m]) / rate(http_server_duration_count[5m])

# P95レイテンシー
histogram_quantile(0.95, rate(http_server_duration_bucket[5m]))

# エンドポイント別リクエスト数
sum by (http_route) (rate(http_server_duration_count[1m]))

# エラーレート
rate(http_server_duration_count{http_status_code=~"5.."}[1m])
```

## 📊 推奨構成の比較

### 最小構成（メトリクスなし）

```
✅ 推奨度: ⭐⭐⭐⭐⭐
サービス: 5個（app, postgres, otel-collector, tempo, loki, grafana）
複雑度: 低
用途: 最初の学習、トレース/ログ重視
```

### シンプル構成（Prometheus）

```
✅ 推奨度: ⭐⭐⭐⭐
サービス: 6個（上記 + prometheus）
複雑度: 低
用途: ローカル開発、中小規模
```

### 中間構成（Mimir Monolithic）

```
✅ 推奨度: ⭐⭐⭐
サービス: 6個（上記 + mimir）
複雑度: 中
用途: 本番環境の予行演習
```

### フル構成（Mimir Distributed）

```
✅ 推奨度: ⭐⭐（ローカルには不要）
サービス: 15個以上
複雑度: 高
用途: 大規模本番環境のみ
```

## 🎯 私の推奨

### ステップ 1: まずはシンプルに

```
Tempo + Loki + Grafana（メトリクスなし）
↓
必要に応じて Prometheus 追加
```

### ステップ 2: 必要なら拡張

```
Prometheus で満足
OR
Mimir Monolithic に移行（学習目的）
```

## 実装の決定

<ask_followup_question>
<question>メトリクスの実装方針を決めましょう。どの構成が良いですか？</question>
<follow_up>
<suggest>Prometheus 構成（シンプル、推奨） - 6 サービス、設定簡単</suggest>
<suggest>メトリクスなし（最小構成） - 5 サービス、トレース+ログのみ、後で追加可能</suggest>
<suggest>Mimir Monolithic（学習目的） - 6 サービス、本番環境の予行演習</suggest>
<suggest>とりあえずメトリクスなしで始めて、後から Prometheus を追加</suggest>
</follow_up>
</ask_followup_question>
