"""
Todo CRUD endpoints.
All observability is handled automatically by opentelemetry-instrument.
"""
import logging
from uuid import UUID
from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from app.config import settings
from app.database import get_db
from app.models.todo import Todo
from app.schemas.todo import TodoCreate, TodoUpdate, TodoResponse


router = APIRouter()
SERVICE_NAME = settings.service_name
logger = logging.getLogger(SERVICE_NAME)


@router.get("/", response_model=List[TodoResponse])
async def get_todos(
    skip: int = 0,
    limit: int = 100,
    db: AsyncSession = Depends(get_db)
) -> List[Todo]:
    """
    Retrieve all todos with pagination.

    - **skip**: Number of records to skip (default: 0)
    - **limit**: Maximum number of records to return (default: 100)
    """
    result = await db.execute(
        select(Todo)
        .order_by(Todo.created_at.desc())
        .offset(skip)
        .limit(limit)
    )
    todos = result.scalars().all()
    logger.info(
        "todos listed",
        extra={"endpoint": "GET /api/v1/todos", "count": len(todos), "skip": skip, "limit": limit},
    )
    return list(todos)


@router.get("/{todo_id}", response_model=TodoResponse)
async def get_todo(
    todo_id: UUID,
    db: AsyncSession = Depends(get_db)
) -> Todo:
    """
    Retrieve a specific todo by ID.

    - **todo_id**: UUID of the todo to retrieve
    """
    result = await db.execute(
        select(Todo).where(Todo.id == todo_id)
    )
    todo = result.scalar_one_or_none()

    if todo is None:
        logger.info(
            "todo not found",
            extra={"endpoint": "GET /api/v1/todos/{id}", "todo_id": str(todo_id), "status": 404},
        )
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Todo with id {todo_id} not found"
        )

    logger.info(
        "todo retrieved",
        extra={"endpoint": "GET /api/v1/todos/{id}", "todo_id": str(todo_id), "status": 200},
    )
    return todo


@router.post("/", response_model=TodoResponse, status_code=status.HTTP_201_CREATED)
async def create_todo(
    todo_data: TodoCreate,
    db: AsyncSession = Depends(get_db)
) -> Todo:
    """
    Create a new todo.

    - **title**: Todo title (required, 1-200 characters)
    - **description**: Todo description (optional)
    - **completed**: Completion status (default: false)
    """
    todo = Todo(**todo_data.model_dump())
    db.add(todo)
    await db.flush()
    await db.refresh(todo)

    logger.info(
        "todo created",
        extra={
            "endpoint": "POST /api/v1/todos",
            "todo_id": str(todo.id),
            "title": todo.title,
            "status": 201,
        },
    )
    return todo


@router.put("/{todo_id}", response_model=TodoResponse)
async def update_todo(
    todo_id: UUID,
    todo_data: TodoUpdate,
    db: AsyncSession = Depends(get_db)
) -> Todo:
    """
    Update an existing todo.

    - **todo_id**: UUID of the todo to update
    - **title**: New title (optional)
    - **description**: New description (optional)
    - **completed**: New completion status (optional)
    """
    result = await db.execute(
        select(Todo).where(Todo.id == todo_id)
    )
    todo = result.scalar_one_or_none()

    if todo is None:
        logger.info(
            "todo not found for update",
            extra={"endpoint": "PUT /api/v1/todos/{id}", "todo_id": str(todo_id), "status": 404},
        )
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Todo with id {todo_id} not found"
        )

    # Update only provided fields
    update_data = todo_data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(todo, field, value)

    await db.flush()
    await db.refresh(todo)

    logger.info(
        "todo updated",
        extra={
            "endpoint": "PUT /api/v1/todos/{id}",
            "todo_id": str(todo_id),
            "updated_fields": list(update_data.keys()),
            "status": 200,
        },
    )
    return todo


@router.delete("/{todo_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_todo(
    todo_id: UUID,
    db: AsyncSession = Depends(get_db)
) -> None:
    """
    Delete a todo.

    - **todo_id**: UUID of the todo to delete
    """
    result = await db.execute(
        select(Todo).where(Todo.id == todo_id)
    )
    todo = result.scalar_one_or_none()

    if todo is None:
        logger.info(
            "todo not found for delete",
            extra={"endpoint": "DELETE /api/v1/todos/{id}", "todo_id": str(todo_id), "status": 404},
        )
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Todo with id {todo_id} not found"
        )

    await db.delete(todo)
    await db.flush()
    logger.info(
        "todo deleted",
        extra={"endpoint": "DELETE /api/v1/todos/{id}", "todo_id": str(todo_id), "status": 204},
    )
