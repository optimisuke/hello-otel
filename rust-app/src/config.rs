use std::env;

#[derive(Debug, Clone)]
pub struct Config {
    pub host: String,
    pub port: u16,
    pub database_url: String,
    pub service_name: String,
}

impl Config {
    pub fn from_env() -> Result<Self, env::VarError> {
        let host = env::var("HOST").unwrap_or_else(|_| "0.0.0.0".to_string());
        let port = env::var("PORT")
            .ok()
            .and_then(|p| p.parse::<u16>().ok())
            .unwrap_or(3003);
        let database_url = env::var("DATABASE_URL")?;
        let service_name = env::var("SERVICE_NAME").unwrap_or_else(|_| "todo-api-rust".to_string());

        Ok(Self {
            host,
            port,
            database_url,
            service_name,
        })
    }
}
