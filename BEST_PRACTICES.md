# ベストプラクティスと実装ガイドライン

## コーディング規約

### Python スタイルガイド

このプロジェクトは[PEP 8](https://pep8.org/)に準拠します。

```python
# ✅ Good
async def get_todo_by_id(todo_id: UUID, db: AsyncSession) -> Todo:
    """特定のTodoを取得する"""
    result = await db.execute(
        select(Todo).where(Todo.id == todo_id)
    )
    return result.scalar_one_or_none()

# ❌ Bad
async def getTodoById(todoId,db):
    result=await db.execute(select(Todo).where(Todo.id==todoId))
    return result.scalar_one_or_none()
```

### 型ヒント

すべての関数に型ヒントを付けます。

```python
from typing import Optional, List
from uuid import UUID

# ✅ Good
async def create_todo(
    todo_ TodoCreate,
    db: AsyncSession
) -> Todo:
    pass

# ❌ Bad
async def create_todo(todo_data, db):
    pass
```

### 非同期処理

FastAPI では非同期処理を活用します。

```python
# ✅ Good - 非同期I/O
async def get_todos(db: AsyncSession) -> List[Todo]:
    result = await db.execute(select(Todo))
    return result.scalars().all()

# ❌ Bad - 同期的な処理
def get_todos(db: Session) -> List[Todo]:
    return db.query(Todo).all()
```

## OpenTelemetry 実装パターン

### 1. トレースのベストプラクティス

#### カスタムスパンの追加

```python
from opentelemetry import trace
from opentelemetry.trace import Status, StatusCode

tracer = trace.get_tracer(__name__)

async def complex_business_logic(todo_id: UUID):
    # 親スパン
    with tracer.start_as_current_span("complex_business_logic") as span:
        span.set_attribute("todo.id", str(todo_id))

        try:
            # 子スパン1
            with tracer.start_as_current_span("validate_todo"):
                # バリデーション処理
                span.add_event("Validation completed")

            # 子スパン2
            with tracer.start_as_current_span("process_todo"):
                # ビジネスロジック
                span.add_event("Processing completed")

            span.set_status(Status(StatusCode.OK))

        except Exception as e:
            span.set_status(Status(StatusCode.ERROR, str(e)))
            span.record_exception(e)
            raise
```

#### スパン属性の標準化

```python
# セマンティック規約に従う
span.set_attribute("http.method", "POST")
span.set_attribute("http.url", "/api/v1/todos")
span.set_attribute("http.status_code", 201)
span.set_attribute("db.system", "postgresql")
span.set_attribute("db.operation", "INSERT")

# カスタム属性
span.set_attribute("todo.title", title)
span.set_attribute("todo.completed", completed)
span.set_attribute("user.action", "create")
```

### 2. メトリクスのベストプラクティス

#### カウンターの使用

```python
from opentelemetry import metrics

meter = metrics.get_meter(__name__)

# カウンターの作成
request_counter = meter.create_counter(
    name="api.requests.total",
    description="Total number of API requests",
    unit="1"
)

# エンドポイントごとにカウント
@router.post("/todos")
async def create_todo(todo: TodoCreate):
    request_counter.add(1, {
        "method": "POST",
        "endpoint": "/api/v1/todos",
        "status": "success"
    })
```

#### ヒストグラムの使用

```python
# レイテンシー測定
latency_histogram = meter.create_histogram(
    name="api.request.duration",
    description="API request duration in milliseconds",
    unit="ms"
)

import time

async def timed_operation():
    start_time = time.time()
    try:
        # 処理
        pass
    finally:
        duration = (time.time() - start_time) * 1000
        latency_histogram.record(duration, {
            "endpoint": "/api/v1/todos",
            "method": "GET"
        })
```

#### ゲージの使用

```python
# 現在の状態を表す
active_connections = meter.create_observable_gauge(
    name="db.connections.active",
    callbacks=[lambda options: get_active_connections()],
    description="Number of active database connections",
    unit="1"
)

def get_active_connections() -> int:
    # データベース接続プールから取得
    return engine.pool.size() - engine.pool.overflow()
```

### 3. ログのベストプラクティス

#### 構造化ログ

```python
import logging
import json
from opentelemetry import trace

logger = logging.getLogger(__name__)

def log_with_context(message: str, level: str = "INFO", **kwargs):
    """トレースコンテキスト付きログ"""
    span = trace.get_current_span()
    span_context = span.get_span_context()

    log_data = {
        "message": message,
        "trace_id": format(span_context.trace_id, "032x"),
        "span_id": format(span_context.span_id, "016x"),
        "timestamp": datetime.utcnow().isoformat(),
        **kwargs
    }

    if level == "ERROR":
        logger.error(json.dumps(log_data))
    elif level == "WARNING":
        logger.warning(json.dumps(log_data))
    else:
        logger.info(json.dumps(log_data))

# 使用例
log_with_context(
    "Todo created successfully",
    level="INFO",
    todo_id=str(todo.id),
    user_action="create"
)
```

## エラーハンドリング

### 標準的なエラーハンドリングパターン

```python
from fastapi import HTTPException, status
from sqlalchemy.exc import IntegrityError

@router.post("/todos", response_model=TodoResponse, status_code=201)
async def create_todo(
    todo: TodoCreate,
    db: AsyncSession = Depends(get_db)
):
    span = trace.get_current_span()

    try:
        # バリデーション
        if not todo.title.strip():
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Title cannot be empty"
            )

        # Todo作成
        new_todo = Todo(**todo.dict())
        db.add(new_todo)
        await db.commit()
        await db.refresh(new_todo)

        span.set_status(Status(StatusCode.OK))
        span.add_event("Todo created successfully")

        log_with_context(
            "Todo created",
            todo_id=str(new_todo.id),
            title=new_todo.title
        )

        return new_todo

    except IntegrityError as e:
        await db.rollback()
        span.set_status(Status(StatusCode.ERROR, "Database integrity error"))
        span.record_exception(e)

        log_with_context(
            "Failed to create todo - integrity error",
            level="ERROR",
            error=str(e)
        )

        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail="Todo already exists"
        )

    except Exception as e:
        await db.rollback()
        span.set_status(Status(StatusCode.ERROR, str(e)))
        span.record_exception(e)

        log_with_context(
            "Unexpected error creating todo",
            level="ERROR",
            error=str(e),
            error_type=type(e).__name__
        )

        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Internal server error"
        )
```

## データベース操作

### トランザクション管理

```python
from sqlalchemy.ext.asyncio import AsyncSession
from contextlib import asynccontextmanager

@asynccontextmanager
async def get_db_transaction():
    """トランザクション付きDBセッション"""
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()

# 使用例
async def complex_operation():
    async with get_db_transaction() as db:
        # 複数の操作
        todo1 = Todo(title="Task 1")
        todo2 = Todo(title="Task 2")
        db.add_all([todo1, todo2])
        # コミットは自動的に実行される
```

### N+1 問題の回避

```python
from sqlalchemy.orm import selectinload

# ❌ Bad - N+1問題
async def get_todos_with_tags():
    todos = await db.execute(select(Todo))
    for todo in todos.scalars():
        # 各Todoごとにクエリが実行される
        tags = await db.execute(
            select(Tag).where(Tag.todo_id == todo.id)
        )

# ✅ Good - Eager Loading
async def get_todos_with_tags():
    result = await db.execute(
        select(Todo).options(selectinload(Todo.tags))
    )
    todos = result.scalars().all()
    # タグも一緒に取得済み
```

### インデックスの活用

```python
from sqlalchemy import Index

class Todo(Base):
    __tablename__ = "todos"

    id = Column(UUID(as_uuid=True), primary_key=True)
    title = Column(String(200), nullable=False)
    completed = Column(Boolean, default=False)
    created_at = Column(DateTime(timezone=True), server_default=func.now())

    # インデックス定義
    __table_args__ = (
        Index('ix_todos_completed', 'completed'),
        Index('ix_todos_created_at', 'created_at'),
        Index('ix_todos_title_completed', 'title', 'completed'),
    )
```

## パフォーマンス最適化

### 1. 接続プーリング

```python
from sqlalchemy.ext.asyncio import create_async_engine

engine = create_async_engine(
    DATABASE_URL,
    pool_size=10,              # 最小接続数
    max_overflow=20,           # 追加接続数
    pool_timeout=30,           # タイムアウト
    pool_recycle=3600,         # 接続再利用時間
    pool_pre_ping=True,        # 接続チェック
    echo=False
)
```

### 2. キャッシング戦略

```python
from functools import lru_cache
from typing import Optional
import asyncio

# メモリキャッシュ
_cache: dict = {}
_cache_ttl: dict = {}

async def get_todo_cached(todo_id: UUID, ttl: int = 300) -> Optional[Todo]:
    """TTL付きキャッシュ"""
    cache_key = f"todo:{todo_id}"

    # キャッシュチェック
    if cache_key in _cache:
        if time.time() < _cache_ttl.get(cache_key, 0):
            return _cache[cache_key]

    # DBから取得
    todo = await get_todo_from_db(todo_id)

    # キャッシュに保存
    if todo:
        _cache[cache_key] = todo
        _cache_ttl[cache_key] = time.time() + ttl

    return todo
```

### 3. バルク操作

```python
# ❌ Bad - 個別にINSERT
for item in items:
    todo = Todo(**item)
    db.add(todo)
    await db.commit()

# ✅ Good - バルクINSERT
todos = [Todo(**item) for item in items]
db.add_all(todos)
await db.commit()
```

## セキュリティ

### 1. SQL インジェクション対策

```python
# ✅ Good - パラメータ化クエリ
async def search_todos(keyword: str):
    result = await db.execute(
        select(Todo).where(Todo.title.ilike(f"%{keyword}%"))
    )
    return result.scalars().all()

# ❌ Bad - 文字列結合（脆弱）
async def search_todos_bad(keyword: str):
    query = f"SELECT * FROM todos WHERE title LIKE '%{keyword}%'"
    result = await db.execute(text(query))
```

### 2. 入力バリデーション

```python
from pydantic import BaseModel, Field, validator

class TodoCreate(BaseModel):
    title: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=5000)

    @validator('title')
    def title_must_not_be_empty(cls, v):
        if not v or not v.strip():
            raise ValueError('Title cannot be empty')
        return v.strip()

    @validator('description')
    def sanitize_description(cls, v):
        if v:
            # HTMLタグの除去など
            return v.strip()
        return v
```

### 3. レート制限

```python
from fastapi import Request
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

@app.post("/todos")
@limiter.limit("10/minute")
async def create_todo(request: Request, todo: TodoCreate):
    pass
```

## テスト戦略

### 1. ユニットテスト

```python
import pytest
from unittest.mock import AsyncMock, patch

@pytest.mark.asyncio
async def test_create_todo():
    # Arrange
    mock_db = AsyncMock()
    todo_data = TodoCreate(
        title="Test Todo",
        description="Test Description"
    )

    # Act
    result = await create_todo(todo_data, mock_db)

    # Assert
    assert result.title == "Test Todo"
    mock_db.add.assert_called_once()
    mock_db.commit.assert_called_once()
```

### 2. 統合テス
