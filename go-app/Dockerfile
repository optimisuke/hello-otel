FROM golang:1.25-alpine AS builder
ARG OTEL_AGENT_URL="https://github.com/alibaba/loongsuite-go-agent/releases/latest/download/otel-linux-amd64"

WORKDIR /app

RUN apk add --no-cache ca-certificates curl
COPY go.mod go.sum ./
RUN go mod download
# Install Loongsuite Go Agent CLI (otel) for build-time auto-instrumentation.
RUN curl -fsSL "${OTEL_AGENT_URL}" -o /usr/local/bin/otel && chmod +x /usr/local/bin/otel

COPY . .
RUN CGO_ENABLED=0 GOOS=linux otel go build -o /bin/server ./cmd/server

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /bin/server /server
USER nonroot:nonroot
EXPOSE 3002
ENTRYPOINT ["/server"]
