use sqlx::PgPool;
use uuid::Uuid;

use crate::todo::model::{CreateTodoRequest, Todo, UpdateTodoRequest};

#[derive(Debug, thiserror::Error)]
pub enum RepoError {
    #[error("todo not found")]
    NotFound,
    #[error("no fields to update")]
    NoFieldsToUpdate,
    #[error(transparent)]
    Sqlx(#[from] sqlx::Error),
}

pub async fn list(pool: &PgPool, skip: i64, limit: i64) -> Result<Vec<Todo>, RepoError> {
    let rows = sqlx::query_as::<_, Todo>(
        r#"
        SELECT id, title, description, completed, created_at, updated_at
        FROM todos
        ORDER BY created_at DESC
        OFFSET $1 LIMIT $2
        "#,
    )
    .bind(skip)
    .bind(limit)
    .fetch_all(pool)
    .await?;

    Ok(rows)
}

pub async fn get(pool: &PgPool, id: Uuid) -> Result<Todo, RepoError> {
    let row = sqlx::query_as::<_, Todo>(
        r#"
        SELECT id, title, description, completed, created_at, updated_at
        FROM todos
        WHERE id = $1
        "#,
    )
    .bind(id)
    .fetch_optional(pool)
    .await?;

    row.ok_or(RepoError::NotFound)
}

pub async fn create(pool: &PgPool, payload: CreateTodoRequest) -> Result<Todo, RepoError> {
    let id = Uuid::new_v4();
    let completed = payload.completed.unwrap_or(false);

    let row = sqlx::query_as::<_, Todo>(
        r#"
        INSERT INTO todos (id, title, description, completed)
        VALUES ($1, $2, $3, $4)
        RETURNING id, title, description, completed, created_at, updated_at
        "#,
    )
    .bind(id)
    .bind(payload.title)
    .bind(payload.description)
    .bind(completed)
    .fetch_one(pool)
    .await?;

    Ok(row)
}

pub async fn update(
    pool: &PgPool,
    id: Uuid,
    payload: UpdateTodoRequest,
) -> Result<Todo, RepoError> {
    let has_any_field =
        payload.title.is_some() || payload.description.is_some() || payload.completed.is_some();
    if !has_any_field {
        return Err(RepoError::NoFieldsToUpdate);
    }

    let mut qb = sqlx::QueryBuilder::new("UPDATE todos SET ");
    let mut separated = qb.separated(", ");

    if let Some(title) = payload.title {
        separated.push("title = ").push_bind(title);
    }
    if let Some(description) = payload.description {
        separated.push("description = ").push_bind(description);
    }
    if let Some(completed) = payload.completed {
        separated.push("completed = ").push_bind(completed);
    }

    separated.push("updated_at = NOW()");

    qb.push(" WHERE id = ").push_bind(id);
    qb.push(" RETURNING id, title, description, completed, created_at, updated_at");

    let row = qb.build_query_as::<Todo>().fetch_optional(pool).await?;

    row.ok_or(RepoError::NotFound)
}

pub async fn delete(pool: &PgPool, id: Uuid) -> Result<(), RepoError> {
    let res = sqlx::query(r#"DELETE FROM todos WHERE id = $1"#)
        .bind(id)
        .execute(pool)
        .await?;

    if res.rows_affected() == 0 {
        return Err(RepoError::NotFound);
    }

    Ok(())
}
