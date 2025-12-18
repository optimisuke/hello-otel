use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::json;
use uuid::Uuid;

use crate::todo::{
    model::{CreateTodoRequest, Pagination, UpdateTodoRequest},
    repo::{self, RepoError},
    AppState,
};

#[derive(Debug, thiserror::Error)]
pub enum ApiError {
    #[error("{0}")]
    BadRequest(String),
    #[error(transparent)]
    Repo(#[from] RepoError),
}

impl IntoResponse for ApiError {
    fn into_response(self) -> Response {
        let (status, message) = match &self {
            ApiError::BadRequest(msg) => (StatusCode::BAD_REQUEST, msg.clone()),
            ApiError::Repo(RepoError::NotFound) => (StatusCode::NOT_FOUND, "not found".to_string()),
            ApiError::Repo(RepoError::NoFieldsToUpdate) => {
                (StatusCode::BAD_REQUEST, "no fields to update".to_string())
            }
            ApiError::Repo(RepoError::Sqlx(_)) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                "internal server error".to_string(),
            ),
        };

        (status, Json(json!({ "error": message }))).into_response()
    }
}

pub async fn health(State(state): State<AppState>) -> impl IntoResponse {
    (
        StatusCode::OK,
        Json(json!({
            "status": "healthy",
            "service": state.service_name,
        })),
    )
}

pub async fn list(
    State(state): State<AppState>,
    Query(params): Query<Pagination>,
) -> Result<impl IntoResponse, ApiError> {
    let skip = params.skip.unwrap_or(0).max(0);
    let limit = params.limit.unwrap_or(100).clamp(1, 1000);

    let todos = repo::list(&state.pool, skip, limit).await?;
    Ok((StatusCode::OK, Json(todos)))
}

pub async fn get(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, ApiError> {
    let todo = repo::get(&state.pool, id).await?;
    Ok((StatusCode::OK, Json(todo)))
}

pub async fn create(
    State(state): State<AppState>,
    Json(mut payload): Json<CreateTodoRequest>,
) -> Result<impl IntoResponse, ApiError> {
    payload.title = payload.title.trim().to_string();

    validate_create(&payload)?;

    let todo = repo::create(&state.pool, payload).await?;
    Ok((StatusCode::CREATED, Json(todo)))
}

pub async fn update(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(mut payload): Json<UpdateTodoRequest>,
) -> Result<impl IntoResponse, ApiError> {
    if let Some(title) = payload.title.as_mut() {
        *title = title.trim().to_string();
    }

    validate_update(&payload)?;

    let todo = repo::update(&state.pool, id, payload).await?;
    Ok((StatusCode::OK, Json(todo)))
}

pub async fn delete(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, ApiError> {
    repo::delete(&state.pool, id).await?;
    Ok(StatusCode::NO_CONTENT)
}

fn validate_create(payload: &CreateTodoRequest) -> Result<(), ApiError> {
    if payload.title.is_empty() {
        return Err(ApiError::BadRequest("title is required".to_string()));
    }
    if payload.title.chars().count() > 200 {
        return Err(ApiError::BadRequest(
            "title must be <= 200 chars".to_string(),
        ));
    }
    Ok(())
}

fn validate_update(payload: &UpdateTodoRequest) -> Result<(), ApiError> {
    if payload.title.is_none() && payload.description.is_none() && payload.completed.is_none() {
        return Err(ApiError::BadRequest("no fields to update".to_string()));
    }
    if let Some(title) = &payload.title {
        if title.is_empty() {
            return Err(ApiError::BadRequest("title must not be empty".to_string()));
        }
        if title.chars().count() > 200 {
            return Err(ApiError::BadRequest(
                "title must be <= 200 chars".to_string(),
            ));
        }
    }
    Ok(())
}
