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

## SQL span についての注意

- Beyla は eBPF ベースで HTTP と SQL の span を生成できますが、Rust の場合は generic tracer (言語によらず汎用的なトレーサー) 扱いになることがあり、**HTTP span の子として同じ Trace に SQL span がぶら下がるとは限りません**（SQL 側が別 Trace になるケースがあります）。
- その場合は HTTP の trace を開いて探すより、Tempo の検索で `db.*` / `server.port=5432` / `span.kind=client` などで **SQL span 側から探す**のが早いです。
