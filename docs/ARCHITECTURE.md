# Architecture

This document describes how `chill-institute/cli` is built.

## System Context

```mermaid
graph LR
  User --> CLI["chilly"]
  CLI --> Config["local config store"]
  CLI --> RPC["HTTP RPC client"]
  RPC --> API["chill.institute API"]
```

## Components

| Component | Responsibility | Talks to |
|-----------|----------------|----------|
| Cobra command layer | Parse commands, flags, and output mode | app context, config store, RPC client |
| App context | Share config path, API URL, output mode, and helpers | commands, config store |
| Config store | Persist local auth token and API base URL | filesystem |
| RPC client | Send JSON requests to v4 procedures, attach auth headers, map errors | `chill.institute` API |
| Output renderers | Render pretty or JSON command output | command handlers |

## Command Model

```mermaid
graph TD
  Root["chilly"] --> Auth["auth"]
  Root --> Whoami["whoami"]
  Root --> Settings["settings"]
  Root --> Search["search"]
  Root --> TopMovies["list-top-movies"]
  Root --> Transfer["add-transfer"]
  Root --> User["user"]
```

Current command groups:

| Command | Responsibility |
|---------|----------------|
| `auth` | login/logout and token acquisition |
| `whoami` | verify current auth state |
| `settings` | inspect and update local CLI config |
| `search` | run search against the hosted API |
| `list-top-movies` | fetch top-movies data |
| `add-transfer` | send transfer requests |
| `user` | user-scoped API operations such as settings reads and writes |

## Local State

```mermaid
graph TD
  CLI["command"] --> Store["config.Store"]
  Store --> File["$XDG_CONFIG_HOME/chilly/config.json"]
```

The config store owns:

- API base URL
- auth token

The store normalizes defaults and writes atomically through a temp-file replace flow.
It also keeps the config directory private (`0700`) and the config file private (`0600`).

## Request Flow

```mermaid
sequenceDiagram
  participant User
  participant Command
  participant Store
  participant Client
  participant API

  User->>Command: chilly search ...
  Command->>Store: load config
  Store-->>Command: api_base_url, auth_token
  Command->>Client: call procedure
  Client->>API: POST /v4/{procedure}
  API-->>Client: JSON response or error envelope
  Client-->>Command: typed result / APIError
  Command-->>User: pretty or JSON output
```

## API Client Model

The current client is intentionally lightweight:

- it sends HTTP POST requests directly to `/v4/{procedure}`
- it supports `none` and `user` auth modes
- it adds `X-Request-Id` for tracing
- it parses the shared error envelope into `APIError`

This repo does not yet consume generated RPC bindings directly. It currently uses a manual procedure-oriented client.

## Boundaries

- Local config is the only persistent state in this repo.
- The CLI does not embed backend behavior. It delegates to the hosted API.
- Auth is bearer-token based for user-scoped commands.
