<p align="center">
  <img src="branding/logo-wordmark.svg" alt="Conduit" width="360"/>
</p>

# Conduit — Project Planner

> **One Console. Every Protocol.**
> A single-page map of **what Conduit is**, **how it's built**, and **where it stands today**.
> For the granular, checkbox-level build log see [`TASKS.md`](./TASKS.md); for the visual
> identity and palette see [`THEME.md`](./THEME.md).

---

## 1. What is Conduit?

**Conduit is a universal protocol workbench** — one native desktop application for connecting
to, testing, and inspecting the protocols and data stores an engineer touches day to day.
Think *a REST client, a database IDE, a Kafka tool, a file-transfer commander, and an AI/MCP
tester* unified behind a single themed, keyboard-driven, searchable workspace — with an
embedded, indexed Help system so you never leave the app to understand a feature.

**Core idea:** every protocol is a *connector* implementing one common interface (the SPI)
plus a *view* (a UI tab). The shell, vault, history, help, environment, and cache are shared
infrastructure that every protocol reuses — so adding a protocol is a small, well-defined
unit of work.

Conduit is built **pure-Go, end to end**: one language, one static binary, native widgets via
**Fyne**. No JVM, no bundled browser, no heavyweight framework.

---

## 2. Goals & Principles

| Principle | What it means in practice |
|-----------|---------------------------|
| **One app, many protocols** | REST, WebSocket, SQL, MongoDB, Kafka, MQTT, gRPC, SFTP, … all in one workspace with consistent UX. |
| **Pluggable by design** | A `Connector` interface + a per-protocol package under `internal/protocol/*`. New protocol = new package + a view wired into the shell. |
| **Secure by default** | Secrets live in an AES-256-GCM vault (PBKDF2, 200k iters), never plaintext on disk. |
| **Meaning through colour** | Every protocol family and component has a semantic colour, so you read the workspace at a glance (see [`THEME.md`](./THEME.md)). |
| **Help is built-in** | Searchable, context-sensitive, `F1`-anywhere — authored alongside each feature. |
| **Idiomatic Go, few heavy deps** | Stdlib first (`net/http`, `database/sql`, `crypto/*`, `encoding/json`); well-known libraries only where the stdlib falls short. Single static binary. |
| **Verify on real targets** | Features are validated against live endpoints / containers, not just unit mocks. |

---

## 3. Architecture at a glance

- **Language / runtime:** Go 1.26+, goroutines + channels for concurrency.
- **UI:** **Fyne v2** — pure-Go native widgets + a custom Conduit theme. A light MV separation:
  `view (widgets) → service (connector) → transport`. Background work runs on goroutines; UI
  updates marshal back onto the UI thread so it never blocks.
- **Build:** a **single Go module** with `internal/` packages. One runnable `main` under
  `cmd/conduit`.
- **Persistence:** SQLite + FTS5 (history) via `modernc.org/sqlite` (pure-Go, cgo-free);
  AES-256-GCM encrypted JSON (vault / connection profiles) under `~/.conduit/`.
- **Cache:** bounded LRU regions (DNS, TLS, schema, lag, help, …).
- **Pattern for "add a protocol":** new `internal/protocol/<name>` package implementing
  `Connector` → a `<Name>View` in `internal/ui/views` → wire into the shell (menu + sidebar +
  tab opener) → register a help topic.

### Package map

```
conduit/                            (single Go module: github.com/geekpratyush/conduit)
├── cmd/conduit/                     ← Fyne entry point (the ONLY runnable package)
├── internal/plugin/                 ← Connector SPI, ConnectionConfig, result types, Registry
├── internal/core/                   ← EventBus, AppContext (DI), LRU cache, History store
│                                      (SQLite+FTS5), EnvironmentService (${VAR}/.env/masking)
├── internal/security/               ← CredentialVault (AES-256-GCM/PBKDF2), cert manager
├── internal/protocol/httpc/         ← REST + WebSocket + SSE + GraphQL
├── internal/protocol/grpcc/         ← gRPC (reflection-based, dynamic)
├── internal/protocol/db/            ← database/sql client + on-demand driver registry
├── internal/protocol/mongo/         ← MongoDB client
├── internal/protocol/redis/         ← Redis client
├── internal/protocol/kafka/         ← Kafka (admin/produce/consume)
├── internal/protocol/mqtt/          ← MQTT
├── internal/protocol/rabbitmq/      ← RabbitMQ (AMQP 0.9.1)
├── internal/protocol/jms/           ← JMS-family via STOMP / AMQP 1.0
├── internal/protocol/cloudmsg/      ← AWS SQS + SNS
├── internal/protocol/ldap/          ← LDAP / Active Directory
├── internal/protocol/snmp/          ← SNMP (v1/v2c GET + WALK)
├── internal/protocol/sftp/          ← SFTP
├── internal/protocol/ftp/           ← FTP / FTPS
├── internal/protocol/objstore/      ← S3 / Azure Blob / GCS behind one explorer
├── internal/protocol/ai/            ← MCP client + LLM / agent tester
├── internal/ui/                     ← shell (MainWindow), theme, views, help system
├── branding/                        ← logo + theme mockup (SVG)
├── docs/                            ← architecture notes
└── test-env/                        ← Docker Compose live-test harness
```

### Library choices

| Concern | Library |
|---------|---------|
| UI toolkit | `fyne.io/fyne/v2` |
| HTTP | `net/http` (stdlib) |
| WebSocket | `nhooyr.io/websocket` |
| gRPC | `google.golang.org/grpc` + `jhump/protoreflect` |
| SQL | `database/sql` + `modernc.org/sqlite`, `jackc/pgx`, `go-sql-driver/mysql` |
| Mongo | `go.mongodb.org/mongo-driver` |
| Redis | `redis/go-redis` |
| Kafka | `twmb/franz-go` |
| MQTT | `eclipse/paho.mqtt.golang` |
| RabbitMQ | `rabbitmq/amqp091-go` |
| JMS-family | `go-stomp/stomp` / `Azure/go-amqp` |
| SQS/SNS/S3 | `aws-sdk-go-v2` |
| LDAP | `go-ldap/ldap/v3` |
| SNMP | `gosnmp/gosnmp` |
| SFTP | `pkg/sftp` + `golang.org/x/crypto/ssh` |
| FTP | `jlaffaye/ftp` |
| Azure Blob | `Azure/azure-sdk-for-go/.../azblob` |
| GCS | `cloud.google.com/go/storage` |
| LLM | `anthropics/anthropic-sdk-go` |
| Vault crypto | `crypto/aes` + `crypto/cipher` + `golang.org/x/crypto/pbkdf2` |
| History FTS | SQLite FTS5 via `modernc.org/sqlite` |
| Markdown | `yuin/goldmark` → Fyne `RichText` |

---

## 4. Phase roadmap

| Phase | Theme | Status |
|-------|-------|--------|
| **0** | Project scaffold (module, packages, Fyne shell, core infra) | ✅ Done — shell runs (themes, sidebar, tabs, log, status bar) |
| **1** | Foundation: vault, cert manager, profiles, env vars, history | ✅ Essentially done — all backends built & tested; some UIs first-cut |
| **2** | Help system (built early to guide everything) | ⬜ Not started |
| **3** | HTTP core: REST, WebSocket, SSE, GraphQL | 🟡 REST backend done & tested; **REST view is the next task** |
| **4** | Kafka client (producer/consumer/admin) | ⬜ Not started |
| **5** | Enterprise messaging (MQTT, RabbitMQ, JMS, cloud SQS/SNS) | ⬜ Not started |
| **6** | Advanced HTTP (gRPC, GraphQL depth) | ⬜ Not started |
| **7** | File transfer (SFTP, FTP/FTPS, S3/Azure/GCS) | ⬜ Not started |
| **8** | Databases & enterprise (SQL, Mongo, Redis, LDAP, SNMP) | ⬜ Not started |
| **9** | Monitoring, metrics, code-gen, packaging | ⬜ Not started |

Legend: ✅ done · 🟡 in progress · ⬜ not started

**Overall: the app runs.** The shared core (event bus, cache, env resolver, history, vault,
cert manager, profile store) and the REST connector are implemented and unit-tested, and the
Fyne shell is live with Midnight/Daylight themes, a colour-coded sidebar, tabs, a log panel, and
a vault control. **Next: the full REST request/response view (Phase 3).** See
[`TASKS.md`](./TASKS.md) for the live checkbox tracker and exact resume point.

**Preview the UI without a display:** `go run ./cmd/preview <outdir>` renders both themes to PNG
via Fyne's software renderer. Run the live app with `go run ./cmd/conduit`.

---

## 5. Build order (highest value first)

1. **Phase 0 — scaffold**: module, package skeletons, `Connector` SPI, `AppContext` DI,
   EventBus, config-dir plumbing, and a minimal Fyne shell (menu + sidebar tree + tabbed
   workspace + log panel + status bar).
2. **Phase 1 — foundation**: credential vault (done), env resolver (done), history store
   (done), plus connection-profile store + samples, and the certificate manager.
3. **Phase 2 — help**: Markdown-rendered searchable help + F1 wiring.
4. **Phase 3 — HTTP core**: REST (the flagship view), then WebSocket, SSE, GraphQL — validates
   the whole view/connector/history/env pipeline end-to-end.
5. **Phases 4–8 — protocol breadth**: Kafka, messaging, gRPC, file transfer, databases,
   directory/monitoring — each a package + a view, reusing the shared infra.
6. **Phase 9 — polish**: metrics dashboard, code generation, single-binary packaging per OS.

---

## 6. Key decisions

- **Fyne (pure-Go native)** — one language, one static binary; rich Markdown/diagram/editor
  rendering handled with custom widgets rather than an embedded browser.
- **Single Go module + `internal/` packages** — enforces encapsulation without workspace overhead.
- **`modernc.org/sqlite`** (pure Go, cgo-free) — keeps `CGO_ENABLED=0` static/cross builds
  working; still has FTS5.
- **Hand-rolled `AppContext` DI** — explicit, greppable wiring over a DI framework.
- **franz-go for Kafka** — pure-Go, no cgo/librdkafka.
- **Verify on real targets** — a `test-env/` Docker stack live-verifies protocols that need a
  broker/server (Kafka, Postgres, RabbitMQ, MQTT, SFTP/FTP, LocalStack).

---

## 7. Author & contact

Built and maintained by **Pratyush Ranjan Mishra**.

- **GitHub:** https://github.com/geekpratyush/conduit
- **LinkedIn:** https://www.linkedin.com/in/leadtherightway/

_See [`README.md`](./README.md) for the About / Contact section and quick start._
