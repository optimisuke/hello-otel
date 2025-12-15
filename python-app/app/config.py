"""
Application configuration using pydantic-settings.
Environment variables are automatically loaded from .env file.
"""
from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Application settings loaded from environment variables."""

    # Database
    database_url: str = "postgresql+asyncpg://todouser:todopass@postgres:5432/tododb"

    # Application
    app_host: str = "0.0.0.0"
    app_port: int = 8000
    log_level: str = "INFO"
    service_name: str = Field("todo-api", alias="OTEL_SERVICE_NAME")
    deployment_environment: str = Field(
        "development", alias="DEPLOYMENT_ENVIRONMENT")

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        populate_by_name=True,
    )


# Global settings instance
settings = Settings()
