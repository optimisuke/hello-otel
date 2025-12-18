pub mod handlers;
pub mod model;
pub mod repo;

use sqlx::PgPool;

#[derive(Clone)]
pub struct AppState {
    pub pool: PgPool,
    pub service_name: String,
}

impl AppState {
    pub fn new(pool: PgPool, service_name: String) -> Self {
        Self { pool, service_name }
    }
}
