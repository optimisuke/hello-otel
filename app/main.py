"""
Todo API with Zero-Code Observability.

All observability (tracing, metrics, logs) is handled automatically
by the opentelemetry-instrument command. No imports or code changes needed!
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.routers import todos

# Create FastAPI app - completely clean, no observability code!
app = FastAPI(
    title="Todo API",
    description="Simple Todo API with Automatic OpenTelemetry Instrumentation",
    version="0.1.0",
    docs_url="/docs",
    redoc_url="/redoc"
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


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "message": "Todo API with Grafana OTEL-LGTM",
        "docs": "/docs",
        "health": "/health"
    }


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {
        "status": "healthy",
        "service": "todo-api"
    }


# Note: No OpenTelemetry imports!
# No manual span creation!
# No metrics recording!
# Everything is automatic via opentelemetry-instrument command!
