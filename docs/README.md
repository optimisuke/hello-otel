# Docs Index

維持するドキュメントだけを整理しました。詳細は以下を参照してください。

- 記事: `docs/articles/otel-auto-instrumentation-by-language.md`
- アーキテクチャ: `docs/architecture/ARCHITECTURE.md`, `docs/architecture/FINAL_ARCHITECTURE.md`
- ガイド: `docs/guides/OBSERVABILITY_GUIDE.md`, `docs/guides/MIMIR_GUIDE.md`, `docs/guides/BEST_PRACTICES.md`

古いプラン系とクイックスタート（本READMEと重複）は削除済みです。

サービス別のREADME
- Python (FastAPI): `python-app/README.md`
- Node (Express + TS + Prisma): `node-app/README.md`
- Spring Boot (Java Agent): `spring-app/README.md`
- Go (chi + sqlx + Loongsuite): `go-app/README.md`
- Rust (axum + sqlx): `rust-app/README.md`

## TODO
- eBPF を使った自動計装の検証
- Alibaba OSS の loongsuite-go-agent を使ったビルド時アプローチの調査: https://github.com/alibaba/loongsuite-go-agent
- gRPC や複数サービスをまたがるトレース継続性の確認
