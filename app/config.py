"""
Application configuration using pydantic-settings.
Environment variables are automatically loaded from .env file.
"""
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Database
    database_url: str = "postgresql+asyncpg://todouser:todopass@postgres:5432/tododb"

    # Application
    app_host: str = "0.0.0.0"
    app_port: int = 8000
    log_level: str = "INFO"

    # OpenTelemetry (configured via environment variables, but documented here)
    # OTEL_EXPORTER_OTLP_ENDPOINT: http://lgtm:4317
    # OTEL_SERVICE_NAME: todo-api
    # OTEL_TRACES_EXPORTER: otlp
    # OTEL_METRICS_EXPORTER: otlp
    # OTEL_LOGS_EXPORTER: otlp

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False
    )


# Global settings instance
settings = Settings()
