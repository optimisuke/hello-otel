FROM python:3.11-slim

WORKDIR /app

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

# Copy dependency files
COPY pyproject.toml ./

# Install dependencies
RUN uv pip install --system -r pyproject.toml

# Copy application code
COPY . .

# Expose port
EXPOSE 8000

# Run with OpenTelemetry auto-instrumentation
CMD ["opentelemetry-instrument", \
    "--traces_exporter", "otlp", \
    "--metrics_exporter", "otlp", \
    "--logs_exporter", "otlp", \
    "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]