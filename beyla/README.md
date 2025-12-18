# Beyla (eBPF) for Rust API

Rust の `rust-api` をコード変更なしで計装するために Grafana Beyla を使います。

## 前提（macOS + Rancher Desktop）

- eBPF は **Rancher Desktop の Linux VM カーネル上で動く**ため、Docker Desktop (macOS) の「ホストカーネル上」で動かすのとは条件が違います。
- `docker compose` で `privileged: true` / `pid: host` を使うため、Rancher Desktop 側で特権コンテナが許可されている必要があります。

## 起動

```bash
docker compose up -d postgres collector lgtm rust-api beyla
```

## 設定

- Beyla の設定ファイル: `beyla/beyla.yaml`
  - `open_ports: "3003"` で Rust API（3003）をターゲットにしています
  - Export 先は既存の `collector:4317` (OTLP gRPC)

## よくあるハマりどころ

- Beyla が起動直後にエラーで落ちる:
  - `/sys/kernel/debug` や `/sys/fs/bpf` の mount ができていない / 権限不足の可能性があります
- 何も計装されない:
  - `rust-api` が実際に 3003 で listen しているか確認してください
  - `docker compose logs -f beyla` で discovery のログを確認してください
