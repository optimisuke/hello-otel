# Todo API with OpenTelemetry & Grafana OTEL-LGTM

最もシンプルな構成で完全な観測性を実現した FastAPI Todo アプリケーション。
FastAPI のコードは `python-app/` にまとめ、今後 Node 版 API などを並走させられるようにしています。

## ✨ 特徴

- 🚀 **観測基盤は 4 サービス** - app, postgres, lgtm（統合観測基盤）, otel-collector（spanmetrics）に加え Node 版 API サービスも同梱
- 🎯 **設定ファイル不要** - docker-compose.yml のみ
- 📊 **完全な観測性** - トレース + ログ + メトリクス
- 🧹 **クリーンコード** - アプリに観測性コードゼロ
- ⚡ **すぐ使える** - 起動後即座に Grafana で確認可能
- 🧩 **Node 版も同梱** - Express + TypeScript + Prisma で同じ PostgreSQL を共有
- 🔧 **uv 管理** - 高速な依存関係管理

## 🛠 技術スタック

### アプリケーション

- **FastAPI** - Python ウェブフレームワーク
- **SQLAlchemy** - ORM
- **PostgreSQL** - データベース
- **uv** - Python 依存関係管理
- **Alembic** - データベースマイグレーション

### 観測性（LGTM 統合）

- **Grafana OTEL-LGTM** - オールインワン観測基盤
  - OpenTelemetry Collector
  - Tempo（トレース）
  - Loki（ログ）
  - Mimir（メトリクス）
  - Grafana（可視化）

### 自動計装

- **opentelemetry-instrument** - コマンドライン自動計装
- コード変更不要の完全自動化

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
# 必要に応じて .env を編集
```

### 3. サービスの起動

```bash
docker-compose up -d
```

### 4. データベースマイグレーション

```bash
docker-compose exec app alembic upgrade head
```
Node 版（Express）は同じ `todos` テーブルを利用するため、Prisma のマイグレーションは不要です（クライアント生成のみ）。

### 5. アクセス

| サービス     | URL                        | 説明                   |
| ------------ | -------------------------- | ---------------------- |
| **API (FastAPI)** | http://localhost:8000      | Python 版 Todo API     |
| **API Docs** | http://localhost:8000/docs | Swagger UI             |
| **API (Node)** | http://localhost:3001      | Express + TypeScript 版 Todo API |
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

### Todo 作成

```bash
curl -X POST http://localhost:8000/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, bread, eggs",
    "completed": false
  }'
```

### Todo 一覧取得

```bash
curl http://localhost:8000/api/v1/todos
```

### 特定の Todo 取得

```bash
curl http://localhost:8000/api/v1/todos/{todo_id}
```

### Todo 更新

```bash
curl -X PUT http://localhost:8000/api/v1/todos/{todo_id} \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

### Todo 削除

```bash
curl -X DELETE http://localhost:8000/api/v1/todos/{todo_id}
```

### ヘルスチェック

```bash
curl http://localhost:8000/health
```

## 💻 ローカル開発

Python (FastAPI) 版は `python-app/` 配下で操作します（docker-compose はリポジトリルートで実行）。

### uv のインストール

```bash
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### 依存関係のインストール

```bash
cd python-app
uv sync
```

### 開発サーバー起動

```bash
cd python-app

# 通常起動
uv run uvicorn app.main:app --reload

# OpenTelemetry自動計装付き
uv run opentelemetry-instrument \
  --traces_exporter otlp \
  --metrics_exporter otlp \
  --logs_exporter otlp \
  uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

### データベースマイグレーション

```bash
# 新しいマイグレーション作成
docker-compose exec app alembic revision --autogenerate -m "description"

# マイグレーション適用
docker-compose exec app alembic upgrade head

# ロールバック
docker-compose exec app alembic downgrade -1
```

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
├── collector.yaml               # spanmetrics 用 OTEL Collector 設定
├── docker-compose.yml           # 4サービス構成
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

- [`FINAL_ARCHITECTURE_V2.md`](FINAL_ARCHITECTURE_V2.md) - 最終アーキテクチャ設計
- [`MIMIR_GUIDE.md`](MIMIR_GUIDE.md) - Mimir vs Prometheus 比較
- [`OBSERVABILITY_GUIDE.md`](OBSERVABILITY_GUIDE.md) - OpenTelemetry 技術ガイド
- [`BEST_PRACTICES.md`](BEST_PRACTICES.md) - コーディング規約

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

# Python依存関係追加
uv add <package-name>

# 依存関係同期
uv sync

# テスト実行（今後追加）
uv run pytest
```

## 🌟 重要なポイント

### ✅ アプリケーションコードはクリーン

[`app/main.py`](python-app/app/main.py)には観測性のコードが**一切ありません**：

```python
from fastapi import FastAPI

app = FastAPI(title="Todo API")

@app.get("/")
async def root():
    return {"message": "Hello World"}

# OpenTelemetryのimportなし！
# スパン作成なし！
# メトリクス記録なし！
```

すべて`opentelemetry-instrument`コマンドが自動で行います。

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
