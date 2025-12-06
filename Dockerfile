FROM python:3.11-slim

WORKDIR /app

# Install uv via pip
RUN pip install --no-cache-dir uv

# Copy dependency files
COPY pyproject.toml ./

# Install dependencies using uv
RUN uv pip install --system --no-cache -r pyproject.toml

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