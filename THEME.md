<p align="center">
  <img src="branding/logo-mark.svg" alt="Conduit mark" width="72"/>
</p>

# Conduit — Visual Identity & Theme

> **Design idea:** in Conduit, *colour carries meaning*. Every protocol family and every kind
> of component is painted from a fixed semantic palette, so you read the workspace — the
> connection tree, the tabs, the status pills — at a glance, without reading a single label.

## The imaginary workspace

This is the theme we are building toward — the **Midnight** dark theme, with a colour-coded
connection tree, accent-barred tabs, semantic status pills, and the palette panel:

![Conduit theme mockup](branding/theme-mockup.svg)

---

## 1. Brand mark & logo

The Conduit mark is a **hub-and-spoke "port"**: one central node with three channels flowing
to three endpoints — *one console, every protocol*. It is filled with the signature
**Indigo → Cyan flow** gradient.

| Asset | File | Use |
|-------|------|-----|
| Logo mark | [`branding/logo-mark.svg`](branding/logo-mark.svg) | app icon, favicon, avatar |
| Wordmark | [`branding/logo-wordmark.svg`](branding/logo-wordmark.svg) | headers, README, docs |
| Theme mockup | [`branding/theme-mockup.svg`](branding/theme-mockup.svg) | reference design |

**Signature gradient — "the flow":** `#6366F1` (Indigo) → `#22D3EE` (Cyan). Used for the mark,
primary buttons, and active-selection accents.

---

## 2. Semantic palette — colour = meaning

Every protocol **domain** owns a colour. Connection-tree nodes, tab accent bars, and category
headers all inherit their domain colour, so a Kafka connection is *always* amber and a database
is *always* emerald.

| Domain | Meaning | Colour | Hex |
|--------|---------|--------|-----|
| **HTTP & Web** (REST · WebSocket · SSE · GraphQL · gRPC) | signal / request–response | Signal Blue | `#3B82F6` |
| **Messaging** (Kafka · MQTT · RabbitMQ · JMS · SQS/SNS) | streams / flow | Stream Amber | `#F59E0B` |
| **Databases** (SQL · MongoDB · Redis) | stored data | Data Emerald | `#10B981` |
| **Files & Objects** (SFTP · FTP · S3 · Azure · GCS) | transfer / movement | Transfer Violet | `#8B5CF6` |
| **Directory & Monitoring** (LDAP · SNMP) | structure / observation | Beacon Teal | `#14B8A6` |
| **AI & MCP** (MCP · LLM · Agent) | intelligence | Neural Magenta | `#EC4899` |
| **Security** (Vault · Certificates) | protection | Cipher Gold | `#EAB308` |

## 3. Status palette

| State | Meaning | Colour | Hex |
|-------|---------|--------|-----|
| Success | 2xx / connected / verified | Success Green | `#22C55E` |
| Warning | 3xx / caution / degraded | Warning Amber | `#F59E0B` |
| Error | 4xx–5xx / failed / danger | Danger Rose | `#F43F5E` |
| Info | timing / neutral metrics | Info Sky | `#38BDF8` |

---

## 4. Neutrals — the two themes

### Midnight (dark, default)

| Token | Role | Hex |
|-------|------|-----|
| Canvas | app background | `#0D1117` |
| Surface | sidebar / panels | `#161B22` |
| Elevated | cards / active tab | `#1F2630` |
| Line | borders / dividers | `#2D3540` |
| Text | primary text | `#E6EDF3` |
| Muted | secondary text | `#8B98A9` |
| Accent | brand indigo / cyan | `#6366F1` / `#22D3EE` |

### Daylight (light)

| Token | Role | Hex |
|-------|------|-----|
| Canvas | app background | `#F7F9FC` |
| Surface | sidebar / panels | `#FFFFFF` |
| Elevated | cards / active tab | `#EEF2F7` |
| Line | borders / dividers | `#D8DFE8` |
| Text | primary text | `#1A2230` |
| Muted | secondary text | `#5A6675` |
| Accent | brand indigo / cyan | `#5457E6` / `#0EA5C4` |

> The light-theme accents are slightly deepened so brand elements keep sufficient contrast on a
> white canvas. Toggle with **Ctrl+Shift+T**; the choice is persisted.

---

## 5. Component treatment — "colour objects with meaning"

- **Connection tree** — each node shows a dot in its **domain colour**; category headers use the
  same colour. You navigate by hue.
- **Tabs** — every workspace tab carries a **3px top accent bar** in its protocol's domain colour.
- **Method / status pills** — HTTP methods and response statuses render as pills: `GET` in Signal
  Blue, `200 OK` in Success Green, `4xx/5xx` in Danger Rose, latency in Info Sky.
- **Buttons** — *primary* = Indigo→Cyan gradient; *secondary* = Surface with a Line border;
  *destructive* = Danger Rose.
- **Vault indicator** — Cipher Gold when locked-with-secrets, Danger Rose when locked empty,
  Success Green when unlocked.
- **Editors / JSON** — syntax accents drawn from the palette (keys in Cyan, strings in Emerald,
  numbers in Amber, booleans in Magenta) for consistency with the brand.

---

## 6. Type & spacing

- **UI font:** Inter (fallback `Segoe UI`, `system-ui`, sans-serif).
- **Monospace:** JetBrains Mono (fallback `ui-monospace`, monospace) — URLs, payloads, code.
- **Radii:** 8px controls · 10–12px cards · 16px window.
- **Rhythm:** 4px base unit; 12–16px gutters.

---

_Author: Pratyush Ranjan Mishra · palette values are the single source of truth for the Fyne
theme in `internal/ui/theme`._
