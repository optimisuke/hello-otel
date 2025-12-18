# Todo API で学ぶ「最小変更」OpenTelemetry：Java/Python/Node/Go/Rust を全部比べる

このリポジトリは、同じ Todo API を **Python(FastAPI) / Node(Express) / Java(Spring Boot) / Go(chi) / Rust(axum)** で実装しつつ、観測基盤（Grafana OTEL-LGTM + OpenTelemetry Collector）に **トレース / ログ / メトリクス** を集約します。

この記事では「アプリに観測コードをなるべく書かない」をテーマに、言語ごとに違う *自動計装の現実解* を並べて整理します。

- Java: `-javaagent` で jar を挿すだけ
- Python: 依存関係に入れて起動コマンドを `opentelemetry-instrument` でラップ（モンキーパッチ）
- Node: SDK を **最初に** 起動するコードを 1 行差し込む（必要ならログも OTEL 側へ）
- Go: Alibaba/Loongsuite の build-time 自動計装で「ビルドコマンドを変える」／ログは file tail で送る
- Rust: Grafana Beyla を eBPF で動かして「アプリに触らず」送る（ただし限界もある）

## 全体アーキテクチャ（最小の観測基盤）

この repo の `docker-compose.yml` は、アプリ群と観測基盤をまとめて起動します。

- **lgtm**: `grafana/otel-lgtm`（Grafana + Tempo + Loki + Mimir）
- **collector**: `otel/opentelemetry-collector-contrib`（spanmetrics と filelog も担当）
- **postgres**: 共有 DB
- **app/node-api/spring-api/go-api/rust-api**: 各言語の Todo API
- **beyla**: Rust を eBPF で自動計装（Linux カーネル必須）

データの流れはざっくりこうです（実体は `docker-compose.yml` / `collector.yaml` / `beyla/beyla.yaml`）。

```mermaid
flowchart LR
  subgraph Apps
    PY[Python FastAPI\nopentelemetry-instrument]
    NODE[Node Express\nNodeSDK auto-instrumentation]
    JAVA[Spring Boot\nopentelemetry-javaagent]
    GO[Go chi\nLoongsuite build-time]
    RUST[Rust axum\n(no SDK)]
    BEYLA[Beyla eBPF]
  end

  COL[otel-collector\nOTLP + spanmetrics + filelog] -->|OTLP| LGTM[Grafana OTEL-LGTM\nTempo/Loki/Mimir]

  PY -->|OTLP traces/metrics/logs| COL
  NODE -->|OTLP traces/logs| COL
  JAVA -->|OTLP traces/metrics/logs| COL
  GO -->|OTLP traces/metrics| COL
  GO -->|JSON log file| COL
  RUST --> BEYLA -->|OTLP traces/metrics| COL
```

## 1) Java：jar を 1 つ追加して `-javaagent`（いちばん “入れるだけ”）

Spring Boot は **OpenTelemetry Java Agent** が一番ラクです。アプリコードには手を入れず、起動コマンドだけ変えます。

- 実装: `spring-app/Dockerfile`
  - 起動時に `opentelemetry-javaagent.jar` をダウンロード
  - `java -javaagent:/app/opentelemetry-javaagent.jar -jar /app/app.jar`

ポイントは「アプリに依存を入れる」ではなく、**ランタイムに agent を差し込む**ことです。CI/CD でも差し替えしやすく、Spring なら HTTP / JDBC などがまとめて拾えます。

## 2) Python：依存関係に入れて “起動コマンドをラップ”（モンキーパッチ）

Python は `opentelemetry-instrument` を使うと、実行コマンドを *ラップ* する形で自動計装できます。

- 実装: `python-app/Dockerfile`
  - `CMD ["opentelemetry-instrument", "...", "uvicorn", "app.main:app", ...]`

いわゆるモンキーパッチなので、アプリコードは汚さずに済む一方で「**起動の入口を必ずラップする**」のが重要です（gunicorn/uvicorn/k8s の entrypoint 変更など）。

この repo では compose 側でログ相関も有効化しています（例）。

- `docker-compose.yml`（Python app）
  - `OTEL_PYTHON_LOG_CORRELATION=true`
  - `OTEL_PYTHON_LOGGING_AUTO_INSTRUMENTATION_ENABLED=true`

## 3) Node：SDK を “最初に” 起動する（auto-instrumentation の鉄則）

Node は “SDK を先に起動してからライブラリを読み込む” のが重要です。後から `require/import` しても、HTTP などのパッチが間に合わないケースが出ます。

- 実装: `node-app/src/index.ts` で `import './otel';` を最初に実行
- 実装: `node-app/src/otel.ts`
  - `NodeSDK` + `getNodeAutoInstrumentations()`
  - Trace: OTLP gRPC（`@opentelemetry/exporter-trace-otlp-grpc`）
  - Logs: OTLP HTTP（`@opentelemetry/exporter-logs-otlp-http`）

### ログについて（SDK のログ or 既存 logger のまま相関だけ付ける）

この repo の Node 実装は、最小の例として OTEL Logs API を直接呼んでいます（`node-app/src/index.ts` の `otelLogs.getLogger(...).emit(...)`）。

ただ、実運用では `pino` / `winston` など既存ロガーを使い続けたいことが多いです。その場合でも、

- ログに `trace_id` / `span_id` を差し込む（ログ相関）
- 送信は stdout 収集 or OTEL Logs exporter

のどちらか（または両方）で “トレースからログに飛べる” 体験は作れます。

## 4) Go：ビルドを変えて自動計装（Loongsuite）＋ログは filelog で拾う

Go は「アプリに SDK を入れない」系の手として、Alibaba/Loongsuite の **build-time 自動計装**を使っています。

- 実装: `go-app/Dockerfile`
  - `otel go build -o /bin/server ./cmd/server`
  - 生成されたバイナリを distroless に入れて実行

ここがポイントで、**コードの import を増やさずに** HTTP/DB のトレース・メトリクスが出ます（送信先は OTEL 標準の env で指定）。

### ログは「trace_id は入るが OTLP 送信はしない」ことがある

自動計装側がログ相関（trace id の注入）まではしてくれても、**アプリログ自体を OTLP で送ってくれるとは限りません**。この repo はその “現実” を素直に回避していて、

- Go 側は zap の JSON ログを **ファイルにも書く**
  - `docker-compose.yml`（go-api）: `LOG_FILE_PATH=/var/log/go-api/app.log`
  - `go-app/cmd/server/main.go` の `newLogger()` が stdout + file の両方に出力
- Collector 側で file tail して Loki に送る
  - `collector.yaml` の `filelog/go_api` receiver
  - compose で `go_api_logs` volume を **collector に read-only マウント**

「アプリはただログを書く」「収集は collector が責務」という分離にすると、言語差の吸収がかなり楽になります。

## 5) Rust：Beyla(eBPF) で “触らずに” 送る（ただし繋がらないこともある）

Rust は SDK を入れず、Grafana Beyla を **eBPF で動かして** HTTP / SQL を拾います。

- 設定: `beyla/beyla.yaml`
  - `instrumentations: ["http", "sql"]`
  - `discovery.instrument.exe_path: "*/todo-api-rust"`（対象バイナリ名で紐づけ）
- compose: `docker-compose.yml` の `beyla` サービス
  - `privileged: true` / `pid: "host"` / BPF 関連 mount

### ハマりどころ：HTTP と SQL のトレースが繋がらない

eBPF は「プロセス外から見える範囲」で相関を推定するため、環境やライブラリの組み合わせ次第で **HTTP span と SQL span が同一 trace にならない**ことがあります。

この repo のメモとしては、

- まずは “出ているか” と “同じ service.name か” を確認
- Beyla の対応範囲（HTTP サーバ / DB ドライバ / TLS など）に依存する
- コンテナ/カーネル要件（Linux / eBPF / 権限）で挙動が変わる

あたりを前提として割り切るのが良さそうです。

## Collector 側の要点（spanmetrics と filelog）

`collector.yaml` は「OTLP を受ける」だけでなく、ローカル検証で嬉しい 2 つの役割を持たせています。

- **spanmetrics**: トレースから RED 系メトリクスを生成（Tempo → Mimir に“トレース由来メトリクス”を作る）
  - `service.pipelines.traces` が `spanmetrics` にも export
  - `service.pipelines.metrics` が `spanmetrics` を receiver として受けて Mimir に送る
- **filelog（Go のログ）**: Go が書いた JSON ログファイルを tail して Loki へ送る
  - `receivers.filelog/go_api` が `/var/log/go-api/*.log` を読む
  - `docker-compose.yml` の `collector` で `go_api_logs` を `:ro` マウント

## 起動手順（記事用の最小手順）

```bash
docker-compose up -d --build

# 任意: 負荷をかけてデータを作る
docker-compose run --rm --profile testing k6
```

- API: `http://localhost:8000`（Python）, `http://localhost:3001`（Node）, `http://localhost:8080`（Spring）, `http://localhost:3002`（Go）, `http://localhost:3003`（Rust）
- Grafana: `http://localhost:3000`（admin/admin）

### Grafana での確認（最低限）

- **Trace**: Explore → Tempo → Search で `service.name`（例: `todo-api-node`）を指定して検索
- **Logs**: Explore → Loki で `{service_name="todo-api"}` など（見えないときは “Last 5m” に）
- **Metrics**: Explore → Mimir で `rate(spanmetrics_latency_count[5m])` など

## まとめ：自動計装は “正解が1つ” ではない

同じ「自動計装」でも、実際は言語ごとに勝ち筋が違います。

- **Java**: agent が最強（入れるだけ）
- **Python**: 起動ラップが最短（ただし入口の管理が重要）
- **Node**: SDK を最初に起動（順序が命）
- **Go**: build-time で割り切る／ログは collector に寄せる
- **Rust**: eBPF は強いが万能ではない（繋がらないケースを知っておく）

### 落とし穴チェックリスト（この repo で実際に効いたやつ）

- **Node**: `otel.ts` は必ず最初に import（`node-app/src/index.ts` の 1 行が肝）
- **Python**: entrypoint を `opentelemetry-instrument` で包む（k8s/Procfile/gunicorn でも同様）
- **Go**: “トレースが出る” と “ログが送れる” は別問題（filelog で収集に寄せると安定）
- **Beyla**: Linux + eBPF + 権限が前提（macOS では Linux VM 上で動かす想定）
- **Collector**: 受信プロトコル（gRPC/HTTP）と endpoint を混ぜると詰みやすい（`4317` と `4318` を意識）

次にやるなら、各言語で “ログ相関を揃える” 方針（OTLP logs で統一するか、stdout/file 収集で統一するか）を決めると、ダッシュボードと運用が一気に楽になります。
