# Python Todo API

FastAPI で実装した Todo API 一式です。スタック全体の使い方や観測性の概要は `../README.md` を参照してください。

## クイックスタート（FastAPI サービス）

```bash
# 環境変数
cp .env.example .env

# 依存インストール（uv 推奨）
uv sync

# サービス起動（リポジトリルートで）
docker-compose up -d

# DB マイグレーション
docker-compose exec app alembic upgrade head
```

## API 使用例

- 作成  
  `curl -X POST http://localhost:8000/api/v1/todos/ -H "Content-Type: application/json" -d '{"title":"Buy groceries","description":"Milk, bread, eggs","completed":false}'`
- 一覧  
  `curl http://localhost:8000/api/v1/todos/`
- 取得  
  `curl http://localhost:8000/api/v1/todos/{todo_id}`
- 更新  
  `curl -X PUT http://localhost:8000/api/v1/todos/{todo_id} -H "Content-Type: application/json" -d '{"completed":true}'`
- 削除  
  `curl -X DELETE http://localhost:8000/api/v1/todos/{todo_id}`
- ヘルス  
  `curl http://localhost:8000/health`

## ローカル開発

```bash
# uv インストール（未導入なら）
curl -LsSf https://astral.sh/uv/install.sh | sh

cd python-app
uv sync

# 通常起動
uv run uvicorn app.main:app --reload

# OpenTelemetry 自動計装付き
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

## 開発コマンドメモ

```bash
# コンテナ起動
docker-compose up -d

# ログ監視
docker-compose logs -f app

# シェル接続
docker-compose exec app /bin/bash

# Python 依存追加
uv add <package-name>

# テスト（必要なら）
uv run pytest
```

## コードの観測性について

[`app/main.py`](app/main.py) には観測性のコードはありません。`opentelemetry-instrument` が自動でトレース/メトリクス/ログを付与します。
