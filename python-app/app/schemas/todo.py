"""
Pydantic schemas for Todo model.
Used for request validation and response serialization.
"""
from datetime import datetime
from uuid import UUID
from typing import Optional
from pydantic import BaseModel, Field, ConfigDict


class TodoBase(BaseModel):
    """Base Todo schema with common fields."""

    title: str = Field(..., min_length=1, max_length=200,
                       description="Todo title")
    description: Optional[str] = Field(None, description="Todo description")
    completed: bool = Field(default=False, description="Completion status")


class TodoCreate(TodoBase):
    """Schema for creating a new Todo."""
    pass


class TodoUpdate(BaseModel):
    """Schema for updating an existing Todo. All fields are optional."""

    title: Optional[str] = Field(None, min_length=1, max_length=200)
    description: Optional[str] = None
    completed: Optional[bool] = None


class TodoResponse(TodoBase):
    """Schema for Todo responses."""

    id: UUID
    created_at: datetime
    updated_at: datetime

    model_config = ConfigDict(from_attributes=True)
