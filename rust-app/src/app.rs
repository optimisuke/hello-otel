use axum::{routing::get, Router};
use sqlx::PgPool;
use tower_http::cors::{Any, CorsLayer};

use crate::todo;

pub fn router(pool: PgPool, service_name: String) -> Router {
    Router::new()
        .route("/health", get(todo::handlers::health))
        .route(
            "/api/v1/todos",
            get(todo::handlers::list).post(todo::handlers::create),
        )
        .route(
            "/api/v1/todos/:id",
            get(todo::handlers::get)
                .put(todo::handlers::update)
                .delete(todo::handlers::delete),
        )
        .route(
            "/api/v1/todos/",
            get(todo::handlers::list).post(todo::handlers::create),
        )
        .route(
            "/api/v1/todos/:id/",
            get(todo::handlers::get)
                .put(todo::handlers::update)
                .delete(todo::handlers::delete),
        )
        .layer(
            CorsLayer::new()
                .allow_origin(Any)
                .allow_methods(Any)
                .allow_headers(Any),
        )
        .with_state(todo::AppState::new(pool, service_name))
}
