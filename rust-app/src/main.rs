mod app;
mod config;
mod db;
mod todo;

use crate::{config::Config, db::connect_pool};
use tracing_subscriber::{fmt, EnvFilter};

#[tokio::main]
async fn main() -> Result<(), anyhow::Error> {
    dotenvy::dotenv().ok();

    fmt()
        .with_env_filter(EnvFilter::try_from_default_env().unwrap_or_else(|_| "info".into()))
        .init();

    let config = Config::from_env()?;
    let pool = connect_pool(&config.database_url).await?;

    let router = app::router(pool, config.service_name.clone());
    let listener = tokio::net::TcpListener::bind((config.host.as_str(), config.port)).await?;

    tracing::info!(
        "listening on http://{}:{}",
        config.host.as_str(),
        config.port
    );
    axum::serve(listener, router).await?;
    Ok(())
}
