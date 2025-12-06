# Todo API with OpenTelemetry & LGTM Stack

OpenTelemetry、Grafana、Tempo、Loki、Prometheus を使用した、完全な観測性を備えた FastAPI Todo アプリケーションです。

## 📋 目次

- [機能](#機能)
- [技術スタック](#技術スタック)
- [前提条件](#前提条件)
- [クイックスタート](#クイックスタート)
- [プロジェクト構造](#プロジェクト構造)
- [API 仕様](#api仕様)
- [観測性](#観測性)
- [開発ガイド](#開発ガイド)
- [トラブルシューティング](#トラブルシューティング)

## ✨ 機能

- ✅ Todo 項目の CRUD 操作（作成・取得・更新・削除）
- 📊 完全な観測性（トレース、メトリクス、ログ）
- 🎯 OpenTelemetry 自動計装
- 📈 Grafana ダッシュボード
- 🗄️ PostgreSQL データベース
- 🐳 Docker Compose による簡単セットアップ
- 🔍 分散トレーシング（Tempo）
- 📝 ログ集約（Loki）
- 📉 メトリクス収集（Prometheus）

## 🛠 技術スタック

### アプリケーション

- **FastAPI** 0.109.0 - 高速 Python ウェブフレームワーク
- **SQLAlchemy** 2.0.25 - ORM
- **PostgreSQL** 16 - データベース
- **Pydantic** 2.5.3 - データバリデーション
- **Alembic** 1.13.1 - データベースマイグレーション

### 観測性（LGTM Stack）

- **Loki** 2.9.3 - ログ集約
- **Grafana** 10.2.3 - 可視化ダッシュボード
- **Tempo** 2.3.1 - 分散トレーシング
- **Prometheus** 2.48.1 - メトリクス収集

### OpenTelemetry

- **OpenTelemetry SDK** 1.22.0
- **OpenTelemetry Collector** 0.91.0
- 自動計装: FastAPI、SQLAlchemy

## 📦 前提条件

- Docker Desktop 4.0+
- Docker Compose 2.0+
- (オプション) Python 3.11+ - ローカル開発用

## 🚀 クイックスタート

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd hello-otel
```

### 2. 環境変数の設定

```bash
cp .env.example .env
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

### 5. 動作確認

#### API ドキュメント

http://localhost:8000/docs

#### Grafana ダッシュボード

http://localhost:3000

- ユーザー名: `admin`
- パスワード: `admin`

#### その他のサービス

- **Tempo**: http://localhost:3200
- **Loki**: http://localhost:3100
- **Prometheus**: http://localhost:9090

## 📁 プロジェクト構造

```
hello-otel/
├── app/
│   ├── __init__.py
│   ├── main.py                     # FastAPIアプリケーション
│   ├── config.py                   # 設定管理
│   ├── database.py                 # DB接続
│   ├── models/
│   │   ├── __init__.py
│   │   └── todo.py                 # SQLAlchemyモデル
│   ├── schemas/
│   │   ├── __init__.py
│   │   └── todo.py                 # Pydanticスキーマ
│   ├── routers/
│   │   ├── __init__.py
│   │   └── todos.py                # Todoエンドポイント
│   ├── services/
│   │   ├── __init__.py
│   │   └── todo_service.py         # ビジネスロジック
│   └── observability/
│       ├── __init__.py
│       ├── tracing.py              # トレース設定
│       ├── metrics.py              # メトリクス設定
│       └── logging.py              # ログ設定
├── alembic/                        # DBマイグレーション
├── grafana/
│   ├── datasources/
│   │   └── datasources.yml         # データソース設定
│   └── dashboards/
│       └── todo-dashboard.json     # ダッシュボード定義
├── otel-collector/
│   └── config.yaml                 # Collector設定
├── docker-compose.yml
├── Dockerfile
├── requirements.txt
├── .env.example
└── README.md
```

## 🔌 API 仕様

### ベース URL

```
http://localhost:8000/api/v1
```

### エンドポイント

#### 1. 全 Todo 取得

```bash
GET /todos

# リクエスト例
curl http://localhost:8000/api/v1/todos

# レスポンス例
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Buy groceries",
    "description": "Milk, bread, eggs",
    "completed": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

#### 2. Todo 取得（ID 指定）

```bash
GET /todos/{todo_id}

# リクエスト例
curl http://localhost:8000/api/v1/todos/550e8400-e29b-41d4-a716-446655440000
```

#### 3. Todo 作成

```bash
POST /todos
Content-Type: application/json

# リクエスト例
curl -X POST http://localhost:8000/api/v1/todos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Buy groceries",
    "description": "Milk, bread, eggs",
    "completed": false
  }'

# レスポンス例
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Buy groceries",
  "description": "Milk, bread, eggs",
  "completed": false,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### 4. Todo 更新

```bash
PUT /todos/{todo_id}
Content-Type: application/json

# リクエスト例
curl -X PUT http://localhost:8000/api/v1/todos/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "completed": true
  }'
```

#### 5. Todo 削除

```bash
DELETE /todos/{todo_id}

# リクエスト例
curl -X DELETE http://localhost:8000/api/v1/todos/550e8400-e29b-41d4-a716-446655440000
```

#### 6. ヘルスチェック

```bash
GET /health

# レスポンス例
{
  "status": "healthy",
  "database": "connected"
}
```

#### 7. メトリクス

```bash
GET /metrics

# Prometheus形式のメトリクスを返却
```

## 📊 観測性

### トレース（Traces）

**Tempo で確認:**

1. Grafana にアクセス: http://localhost:3000
2. 左メニュー > Explore
3. データソース: Tempo
4. Search タブでトレースを検索

**確認できる情報:**

- リクエストの全体フロー
- 各処理のレイテンシー
- データベースクエリの実行時間
- エラースタックトレース

### メトリクス（Metrics）

**Prometheus で確認:**
http://localhost:9090

**主要メトリクス:**

- `http_server_requests_total` - HTTP リクエスト総数
- `http_server_duration_seconds` - リクエストレイテンシー
- `todos_total` - Todo 総数
- `todos_completed_ratio` - 完了率
- `db_query_duration_seconds` - DB クエリ時間

**Grafana ダッシュボード:**

- リクエストレート
- レスポンスタイム（p50, p95, p99）
- エラーレート
- Todo 統計

### ログ（Logs）

**Loki で確認:**

1. Grafana にアクセス
2. 左メニュー > Explore
3. データソース: Loki
4. LogQL クエリで検索

**ログクエリ例:**

```logql
# エラーログのみ
{job="todo-api"} |= "ERROR"

# 特定のTrace IDに関連するログ
{job="todo-api"} |= "trace_id=abc123"

# 特定エンドポイントのログ
{job="todo-api"} | json | endpoint="/api/v1/todos"
```

### 相関分析

トレース、メトリクス、ログは相互に関連付けられています:

1. **トレースからログへ:**

   - Tempo でトレース表示時、`Logs for this span`ボタンクリック
   - 自動的に Loki で関連ログを表示

2. **ログからトレースへ:**

   - Loki のログに含まれる trace_id をクリック
   - 自動的に Tempo でトレースを表示

3. **メトリクスからトレースへ:**
   - Prometheus グラフで異常値をクリック
   - Exemplar からトレースにジャンプ

## 💻 開発ガイド

### ローカル開発セットアップ

```bash
# 仮想環境作成
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 依存関係インストール
pip install -r requirements.txt

# 環境変数設定
export DATABASE_URL="postgresql+asyncpg://todouser:todopass@localhost:5432/tododb"
export OTEL_EXPORTER_OTLP_ENDPOINT="http://localhost:4317"

# データベースのみDocker起動
docker-compose up -d postgres otel-collector tempo loki prometheus grafana

# マイグレーション実行
alembic upgrade head

# アプリケーション起動
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

### 新しいマイグレーションの作成

```bash
# マイグレーションファイル生成
docker-compose exec app alembic revision --autogenerate -m "Add new column"

# マイグレーション適用
docker-compose exec app alembic upgrade head

# ロールバック
docker-compose exec app alembic downgrade -1
```

### テスト実行

```bash
# ユニットテスト
docker-compose exec app pytest tests/

# 統合テスト
docker-compose exec app pytest tests/integration/

# カバレッジ
docker-compose exec app pytest --cov=app tests/
```

### ログ確認

```bash
# 全サービスのログ
docker-compose logs -f

# 特定サービスのログ
docker-compose logs -f app
docker-compose logs -f postgres
docker-compose logs -f otel-collector
```

## 🐛 トラブルシューティング

### 問題: アプリケーションが起動しない

```bash
# コンテナの状態確認
docker-compose ps

# ログ確認
docker-compose logs app

# コンテナ再起動
docker-compose restart app
```

### 問題: データベース接続エラー

```bash
# PostgreSQLの状態確認
docker-compose ps postgres

# データベースログ確認
docker-compose logs postgres

# データベース再起動
docker-compose restart postgres
```

### 問題: トレースが Tempo に表示されない

1. OTel Collector のログ確認:

```bash
docker-compose logs otel-collector
```

2. Collector のエンドポイント確認:

```bash
curl http://localhost:4317
```

3. アプリケーションの環境変数確認:

```bash
docker-compose exec app env | grep OTEL
```

### 問題: Grafana でデータソースに接続できない

1. Grafana 再起動:

```bash
docker-compose restart grafana
```

2. データソース設定確認:

   - Grafana > Configuration > Data sources
   - 各データソースの接続テスト実行

3. ネットワーク確認:

```bash
docker-compose exec grafana ping tempo
docker-compose exec grafana ping loki
docker-compose exec grafana ping prometheus
```

### 問題: ポート競合

```bash
# 使用中のポート確認
lsof -i :8000  # macOS/Linux
netstat -ano | findstr :8000  # Windows

# docker-compose.ymlでポート変更
# 例: "8001:8000"
```

### 全サービスのリセット

```bash
# 全コンテナ停止・削除
docker-compose down

# ボリューム含めて削除
docker-compose down -v

# イメージも削除
docker-compose down --rmi all

# 再構築
docker-compose up -d --build
```

## 📚 参考資料

- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [OpenTelemetry Python](https://opentelemetry.io/docs/instrumentation/python/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)
- [Loki Documentation](https://grafana.com/docs
