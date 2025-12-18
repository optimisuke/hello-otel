import logging
import os
import sys

from openai import OpenAI
from opentelemetry.exporter.otlp.proto.grpc._log_exporter import OTLPLogExporter
from opentelemetry.exporter.otlp.proto.grpc.metric_exporter import OTLPMetricExporter
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from traceloop.sdk import Traceloop
from traceloop.sdk.decorators import task, workflow
from traceloop.sdk.instruments import Instruments


def _require_env(name: str) -> str:
    value = os.getenv(name)
    if value:
        return value
    print(f"Missing required env var: {name}", file=sys.stderr)
    sys.exit(2)


def _normalize_otlp_grpc_endpoint(endpoint: str) -> str:
    if endpoint.startswith("http://"):
        return endpoint.removeprefix("http://")
    if endpoint.startswith("https://"):
        return endpoint.removeprefix("https://")
    return endpoint


def init_traceloop() -> None:
    service_name = os.getenv("OTEL_SERVICE_NAME", "llm-job")
    otlp_endpoint = _normalize_otlp_grpc_endpoint(
        os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://collector:4317")
    )
    deployment_env = os.getenv("DEPLOYMENT_ENVIRONMENT", "development")

    span_exporter = OTLPSpanExporter(endpoint=otlp_endpoint, insecure=True)
    metric_exporter = OTLPMetricExporter(endpoint=otlp_endpoint, insecure=True)
    log_exporter = OTLPLogExporter(endpoint=otlp_endpoint, insecure=True)

    Traceloop.init(
        app_name=service_name,
        exporter=span_exporter,
        metrics_exporter=metric_exporter,
        logging_exporter=log_exporter,
        # job なので短時間で終了する。バッチだと終了時に送信しきれないことがあるため即時送信に寄せる。
        disable_batch=True,
        # Collector/LGTM側で探しやすいように最低限だけ付与
        resource_attributes={
            "service.name": service_name,
            "deployment.environment": deployment_env,
        },
        instruments={Instruments.OPENAI},
    )


@task(name="openai_call")
def call_openai(model: str, prompt: str) -> str:
    client = OpenAI(api_key=_require_env("OPENAI_API_KEY"))
    resp = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": prompt}],
    )
    return resp.choices[0].message.content or ""


@workflow(name="llm_job")
def run_workflow(model: str, prompt: str) -> str:
    text = call_openai(model=model, prompt=prompt)
    return text


def main() -> int:
    init_traceloop()

    model = os.getenv("OPENAI_MODEL", "gpt-4o-mini")
    prompt = os.getenv("OPENAI_PROMPT", "俳句を詠んで OpenTelemetryについて")

    logger = logging.getLogger("llm-job")
    logger.info("Starting LLM job", extra={"openai.model": model})

    text = run_workflow(model=model, prompt=prompt)
    logger.info("LLM job completed", extra={"output.preview": text[:200]})
    print(text)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
