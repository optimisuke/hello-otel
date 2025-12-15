"""Todo API with Zero-Code Observability.

All observability (tracing, metrics, logs) is handled automatically
by the opentelemetry-instrument command. No manual spans/metrics needed.
"""
import logging

from fastapi import FastAPI, Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from app.routers import todos
from app.config import settings

# --- logger 設定 ---
SERVICE_NAME = settings.service_name
logger = logging.getLogger(SERVICE_NAME)
logger.setLevel(settings.log_level.upper())

# Create FastAPI app - completely clean, no observability code!
app = FastAPI(
    title="Todo API",
    description="Simple Todo API with Automatic OpenTelemetry Instrumentation",
    version="0.1.0",
    docs_url="/docs",
    redoc_url="/redoc",
    redirect_slashes=False  # 末尾スラッシュのリダイレクトを無効化
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Configure appropriately for production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(
    todos.router,
    prefix="/api/v1/todos",
    tags=["todos"]
)


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError):
    """Return 400 with validation detail and log the bad payload."""
    logger.warning(
        "request validation failed",
        extra={
            "endpoint": request.url.path,
            "service": SERVICE_NAME,
            "errors": exc.errors(),
        },
    )
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={"detail": exc.errors()},
    )


@app.get("/")
async def root():
    """Root endpoint."""
    logger.info("root called", extra={"endpoint": "/", "service": SERVICE_NAME})
    return {
        "message": "Todo API with Grafana OTEL-LGTM",
        "docs": "/docs",
        "health": "/health"
    }


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    logger.info("health_check called", extra={"endpoint": "/health", "service": SERVICE_NAME})
    return {
        "status": "healthy",
        "service": SERVICE_NAME
    }
