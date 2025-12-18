FROM rust:1.88-slim-bookworm AS builder

WORKDIR /app

COPY Cargo.toml ./
COPY src ./src

RUN cargo build --release

FROM debian:bookworm-slim

RUN apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/target/release/todo-api-rust /app/todo-api-rust

ENV HOST=0.0.0.0
ENV PORT=3003

EXPOSE 3003

CMD ["/app/todo-api-rust"]
