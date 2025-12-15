# クイックスタートガイド

OpenTelemetry + Grafana OTEL-LGTM を使用した Todo アプリケーションをすぐに起動できます。

## 🚀 3 ステップで起動

### 1. サービスの起動

```bash
docker-compose up -d
```

これだけで以下のサービスが起動します：

- **app**: FastAPI アプリケーション (http://localhost:8000)
- **postgres**: PostgreSQL データベース
- **lgtm**: Grafana OTEL-LGTM（統合観測基盤）(http://localhost:3000)

### 2. データベースマイグレーション

初回のみ実行：

```bash
# マイグレーションファイルの作成
docker-compose exec app alembic revision --autogenerate -m "Initial migration"

# マイグレーションの適用
docker-compose exec app alembic upgrade head
```

### 3. 動作確認

#### API 確認

```bash
# ヘルスチェック
curl http://localhost:8000/health

# API ドキュメント
open http://localhost:8000/docs
```

#### Grafana 確認

```bash
# Grafana UI
open http://localhost:3000
```

**ログイン情報:**

- ユーザー名: `admin`
- パスワード: `admin`

## 📊 Todo の操作

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
curl http://localhost:8000/api/v1/todos | jq
```

### Todo の更新

```bash
curl -X PUT http://localhost:8000/api/v1/todos/{todo_id} \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

### Todo の削除

```bash
curl -X DELETE http://localhost:8000/api/v1/todos/{todo_id}
```

## 🔍 観測性の確認

### 1. トレースの確認（Tempo）

Grafana (http://localhost:3000) にアクセス：

1. 左メニュー → **Explore**
2. データソース: **Tempo** を選択
3. **Search** タブでトレースを検索
4. 任意のトレースをクリックして詳細表示

**確認できる情報:**

- HTTP リクエストのフロー
- SQL クエリの実行時間
- エンドポイントごとのレイテンシー
- エラーの詳細

### 2. ログの確認（Loki）

1. 左メニュー → **Explore**
2. データソース: **Loki** を選択
3. LogQL クエリ例:

```logql
# すべてのログ
{service_name="todo-api"}

# エラーログのみ
{service_name="todo-api"} |= "ERROR"

# 特定のTrace IDのログ
{service_name="todo-api"} |= "trace_id=abc123"
```

### 3. メトリクスの確認（Mimir）

1. 左メニュー → **Explore**
2. データソース: **Mimir** を選択
3. PromQL クエリ例:

```promql
# リクエストレート（1分あたり）
rate(http_server_duration_count[1m])

# 平均レスポンスタイム
rate(http_server_duration_sum[5m]) / rate(http_server_duration_count[5m])

# エンドポイント別リクエスト数
sum by (http_route) (rate(http_server_duration_count[1m]))
```

## 🛠 開発時のコマンド

### ログ監視

```bash
# 全サービスのログ
docker-compose logs -f

# アプリのみ
docker-compose logs -f app

# LGTMのみ
docker-compose logs -f lgtm
```

### コンテナに入る

```bash
docker-compose exec app /bin/bash
```

### マイグレーション操作

```bash
# 新しいマイグレーション作成
docker-compose exec app alembic revision --autogenerate -m "Add new column"

# マイグレーション適用
docker-compose exec app alembic upgrade head

# ロールバック
docker-compose exec app alembic downgrade -1

# マイグレーション履歴確認
docker-compose exec app alembic history
```

### 再起動

```bash
# 全サービス再起動
docker-compose restart

# アプリのみ再起動
docker-compose restart app
```

## 🧹 クリーンアップ

### サービス停止

```bash
docker-compose down
```

### データも削除（完全リセット）

```bash
docker-compose down -v
```

### イメージも削除

```bash
docker-compose down --rmi all
```

## 🐛 トラブルシューティング

### アプリが起動しない

```bash
# ログ確認
docker-compose logs app

# コンテナ状態確認
docker-compose ps

# 再ビルド
docker-compose up -d --build app
```

### データベース接続エラー

```bash
# PostgreSQL状態確認
docker-compose ps postgres

# データベースログ確認
docker-compose logs postgres

# PostgreSQL再起動
docker-compose restart postgres
```

### トレースが表示されない

```bash
# 環境変数確認
docker-compose exec app env | grep OTEL

# LGTMログ確認
docker-compose logs lgtm

# アプリ再起動
docker-compose restart app
```

### ポート競合エラー

```bash
# 使用中のポート確認（macOS/Linux）
lsof -i :8000
lsof -i :3000
lsof -i :5432

# Windows
netstat -ano | findstr :8000
```

別のポートを使用する場合は、[`docker-compose.yml`](docker-compose.yml)を編集：

```yaml
services:
  app:
    ports:
      - "8001:8000" # ホスト側ポートを変更
```

## 📚 次のステップ

1. **API をテスト** - Swagger UI (http://localhost:8000/docs) で動作確認
2. **トレース確認** - いくつかリクエストを送信して Grafana でトレースを確認
3. **ダッシュボード作成** - Grafana でカスタムダッシュボードを作成
4. **カスタマイズ** - 必要に応じてコードを拡張

## ✨ 重要なポイント

### アプリケーションコードはクリーン

[`app/main.py`](app/main.py)を見ると、OpenTelemetry のコードが**一切ありません**：

```python
from fastapi import FastAPI

app = FastAPI(title="Todo API")

# OpenTelemetryのimportなし！
# トレース作成コードなし！
# メトリクス記録コードなし！
```

すべて`opentelemetry-instrument`コマンドが自動で行っています（[`Dockerfile`](python-app/Dockerfile)参照）。

### シンプルな構成

わずか**3 つのサービス**だけ：

- `app` - FastAPI
- `postgres` - データベース
- `lgtm` - 統合観測基盤（Tempo + Loki + Mimir + Grafana + OTel Collector）

設定ファイルも最小限で、すぐに使い始められます！

---

**これで完全な観測性を備えた Todo アプリケーションの準備完了です！** 🎉
