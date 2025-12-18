# Rust Todo API (axum + sqlx)

既存の PostgreSQL `todos` テーブルをそのまま利用する Rust 実装です（マイグレーションは行いません）。

## 起動

```bash
export DATABASE_URL=postgresql://todouser:todopass@localhost:5432/tododb
export PORT=3003
cargo run
```

## エンドポイント

- `GET    /health`
- `GET    /api/v1/todos?skip=0&limit=100`
- `GET    /api/v1/todos/{id}`
- `POST   /api/v1/todos`
- `PUT    /api/v1/todos/{id}`
- `DELETE /api/v1/todos/{id}`

