# Conduit — Build Log & Task Tracker

> The **living** checkbox-level tracker for Conduit, the universal protocol workbench.
> Source of truth for what's done, in-progress, and pending. Companion to
> [`PROJECT_PLAN.md`](./PROJECT_PLAN.md) (the one-page map) and [`THEME.md`](./THEME.md)
> (visual identity & palette).

**Legend:** `[x]` done · `[~]` in progress / first cut · `[ ]` not started

**Resume point:** _Phase 0 — wiring the Fyne shell skeleton on top of the completed core services._

---

## Phase 0 — Project scaffold

- [x] Create module `github.com/geekpratyush/conduit` + `internal/` package skeleton
- [x] `PROJECT_PLAN.md`, `TASKS.md`, `THEME.md` authored
- [x] Branding: SVG logo mark, wordmark, and a full theme mockup (`branding/`)
- [x] `internal/plugin`: `Connector` SPI interface, `ConnectionConfig`, result types, `Registry`
- [x] `internal/core`: `EventBus` (typed pub/sub) — with unsubscribe
- [x] `internal/core`: `AppContext` (hand-rolled DI container) + config-dir helper (`~/.conduit`)
- [x] `internal/core`: LRU cache with pre-registered regions + tests
- [ ] `cmd/conduit`: minimal Fyne shell — window, menu bar, sidebar connection tree,
      tabbed workspace, collapsible log panel, status bar, accelerators
- [ ] `internal/ui/theme`: custom Conduit theme (Midnight dark + Daylight light palettes) + toggle
- [ ] App builds and launches (`go run ./cmd/conduit`) showing an empty themed workspace
- [ ] `README.md` (done) + `RUN.md`

## Phase 1 — Foundation (vault, cert manager, profiles, env, history)

- [x] `internal/security`: `CredentialVault` — AES-256-GCM + PBKDF2 (200k iters), encrypted
      JSON at `~/.conduit/vault.enc`; unlock/lock; **round-trip + wrong-password + locked-op tests**
- [x] `internal/core`: `EnvironmentService` — `${VAR}` envs, `.env`, `${VAR:-default}`,
      nested resolution, `\${...}` escape, secret masking; **5 unit tests**
- [x] `internal/core`: `HistoryStore` — SQLite + FTS5 (`modernc.org/sqlite`), favorites,
      replay fetch; **add/FTS-search/favorite/replay test**
- [ ] Vault UI: master-password dialog, auto-lock timer, status-bar lock toggle
- [ ] `internal/core`: connection-profile store → `~/.conduit/connections.json`;
      bundled public sample endpoints (deletable); multi-method auth model
- [ ] Saved-connection secrets stored as vault refs (no plaintext)
- [ ] `internal/security`: certificate manager — generate self-signed RSA/ECDSA, PKCS#10 CSR,
      import/export PEM/DER/PKCS12, keystore persist, expiry watchdog + UI

## Phase 2 — Help system

- [ ] Markdown rendering: `goldmark` → Fyne `RichText` (headings, code, lists, links, tables)
- [ ] 3-pane searchable help dialog (topic tree · content · search) with debounced live search
- [ ] Context-sensitive `F1` anywhere + tips
- [ ] Author initial help topics (overview, vault, env vars, each protocol as built)

## Phase 3 — HTTP core (REST, WebSocket, SSE, GraphQL)

- [ ] `internal/protocol/httpc`: REST connector — all methods, params/headers/body, timing,
      HTTP/2, JSON pretty-print
- [ ] REST auth: Basic / Bearer / API-key / OAuth2 (client-creds + auth-code/PKCE) /
      AWS SigV4 / Digest / HMAC
- [ ] REST view: params/headers/body/auth tabs, colour-coded status, response viewers, history
- [ ] REST: cookie jar, response assertions (Tests tab), waterfall timeline, code-gen
- [ ] WebSocket connector + view (connect/disconnect, timestamped log, send bar)
- [ ] SSE connector + view (`text/event-stream` log, event-type filter) — verify live
- [ ] GraphQL connector + view (query/variables editor, one-click introspection, schema explorer)

## Phase 4 — Kafka

- [ ] `internal/protocol/kafka`: connector (franz-go) — admin, produce, consume
- [ ] Kafka view: topic explorer, produce form, consume table + payload formatter + export
- [ ] Consumer-lag monitor, offset-reset, connect diagnostics
- [ ] (Needs a broker for E2E — see `test-env/`)

## Phase 5 — Enterprise messaging

- [ ] `internal/protocol/mqtt`: connect / subscribe / publish — verify live
- [ ] `internal/protocol/rabbitmq`: declare exchange/queue/binding, publish, consume
- [ ] `internal/protocol/jms`: JMS-family via STOMP / AMQP 1.0
- [ ] `internal/protocol/cloudmsg`: AWS SQS + SNS — verify vs LocalStack

## Phase 6 — Advanced HTTP (gRPC)

- [ ] `internal/protocol/grpcc`: reflection-based discovery + unary invoke (JSON ↔ dynamic
      protobuf) — verify live
- [ ] `.proto` file parser + status-code registry
- [ ] GraphQL subscriptions (streaming)

## Phase 7 — File transfer

- [ ] `internal/protocol/sftp`: remote dir tree browse + file read — verify live
- [ ] `internal/protocol/ftp`: FTP / FTPS — verify live
- [ ] Dual-pane commander view (local↔remote, drag-drop, transfer queue: speed/ETA/pause/
      throttle/recursive/integrity-verify, move, batch-rename, dir-compare + sync, bookmarks)
- [ ] `internal/protocol/objstore`: S3 — verify live (MinIO)
- [ ] objstore: Azure Blob + GCS behind the shared bucket→object explorer

## Phase 8 — Databases & enterprise

- [ ] `internal/protocol/db`: `database/sql` client + on-demand driver registry
      (SQLite/Postgres/MySQL/MariaDB), sortable/filterable result grid, JSON/CSV export
- [ ] SQL: object explorer (db→tables→columns), ER diagram, driver-specific TLS
- [ ] `internal/protocol/mongo`: find/aggregate/explain/CRUD, schema diagram, views, export
- [ ] `internal/protocol/redis`: key browser (typed value rendering) + command console
- [ ] `internal/protocol/ldap`: plain/LDAPS + bind, RFC-4515 search, filter builder,
      entry CRUD, LDIF import/export, lazy DIT tree
- [ ] `internal/protocol/snmp`: v1/v2c GET + WALK, MIB-name resolution, trap receiver

## Phase 8b — AI / MCP

- [ ] `internal/protocol/ai`: MCP client — JSON-RPC 2.0 over HTTP/SSE + stdio, Bearer auth
- [ ] AI / LLM tester — Anthropic Go SDK, token usage
- [ ] AI Agent — MCP tool-calling loop (tool_use → execute → tool_result → repeat), live transcript

## Phase 9 — Monitoring, code-gen, packaging

- [ ] Metrics dashboard (throughput / error-rate / P50-P95-P99 + live chart)
- [ ] Global code-generation SPI
- [ ] Single-binary packaging per OS (`fyne package`) + icon/metadata
- [ ] `test-env/`: Docker Compose live-test harness + gated live integration tests

---

## Progress log

- **2026-07-05** — Project created. Scaffolded the module + package skeleton. Established the
  brand (**Conduit**, "One Console. Every Protocol.") with an original SVG logo mark, wordmark,
  and a full theme mockup; documented the semantic palette in `THEME.md`. Implemented and
  unit-tested the shared core: `Connector` SPI + `Registry`, `EventBus`, `AppContext` DI, the
  LRU cache, the `${VAR}` environment resolver (incl. `.env`, defaults, nesting, escaping,
  masking), the SQLite+FTS5 `HistoryStore`, and the AES-256-GCM `CredentialVault`.
  **All packages build; `go test ./internal/...` is green.** Next: the Fyne shell + theme, then
  the REST view as the first end-to-end protocol.

## Decisions log

- **UI = Fyne v2** — pure-Go native toolkit; one language, one static binary. Rich Markdown/
  diagram/editor rendering handled with custom widgets rather than an embedded browser.
- **Single Go module + `internal/` packages** — encapsulation without workspace overhead.
- **`modernc.org/sqlite`** — cgo-free SQLite (keeps `CGO_ENABLED=0` static builds), still FTS5.
- **franz-go for Kafka** — pure-Go, no cgo/librdkafka.
- **Hand-rolled `AppContext` DI** — explicit, greppable wiring.
- **Semantic colour system** — each protocol family/component has a meaning-bearing colour
  (see `THEME.md`), so the workspace is readable at a glance.
