# Architecture

This document describes the system design, package responsibilities, and the end-to-end connection lifecycle.

## High-Level Overview

EasyVPN CLI is a Go application structured in layers:

```
┌──────────────────────────────────────────────────────────────┐
│                        cmd/easyvpn                           │
│              Cobra commands + interactive menu               │
└────────────────────────────┬─────────────────────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────┐
│                      internal/ui                             │
│           Survey prompts, PTerm spinners/tables              │
└────────────────────────────┬─────────────────────────────────┘
                             │
┌────────────────────────────▼─────────────────────────────────┐
│                      internal/core                           │
│         Manager (connect/disconnect) + Reconciler            │
└──────┬─────────────────────┬─────────────────────┬───────────┘
       │                     │                     │
       ▼                     ▼                     ▼
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ internal/   │    │  internal/api   │    │ internal/       │
│ platform    │    │  Discovery,     │    │ state           │
│ adapters    │    │  Agent, IP      │    │ Store           │
└─────────────┘    └────────┬────────┘    └─────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
       ┌─────────────┐            ┌─────────────┐
       │  Supabase   │            │ Node Agent  │
       │  (REST)     │            │ (HTTP)      │
       └─────────────┘            └─────────────┘
```

## Package Reference

### `cmd/easyvpn`

Application entry point. Defines Cobra commands and wires initialization:

- `initConfig()` — Loads config, runs onboarding if needed, initializes logger and VPN manager
- Commands: `setup`, `connect`, `disconnect`, `status`, `ip`
- Default action (no subcommand): interactive menu loop

### `internal/core`

**`Manager`** (`vpn.go`) — Orchestrates the VPN lifecycle:

| Method | Responsibility |
|--------|----------------|
| `Connect()` | Health check → keygen → provision/rotate peer → apply tunnel → save state |
| `Disconnect()` | Disable kill-switch → destroy tunnel → clear state |
| `GetAdapter()` | Expose platform adapter for setup/status |

**`Reconciler`** (`reconciler.go`) — Truth-first state engine:

| Method | Responsibility |
|--------|----------------|
| `Sync()` | Compare persisted state vs OS reality; heal desync |
| `CheckKillSwitch()` | Re-apply firewall rules if enabled |

### `internal/platform`

Implements `PlatformAdapter` interface (`adapter.go`):

```go
type PlatformAdapter interface {
    Name() string
    IsPrivileged() (bool, error)
    CheckRequirements() []models.EasyVPNError
    InstallDependencies() error
    CreateTunnel(cfg *models.WireGuardConfig) error
    DestroyTunnel() error
    SetKillSwitch(enabled bool, serverIP string) error
    GetStatus() (*TunnelStatus, error)
}
```

`GetAdapter()` selects implementation by `runtime.GOOS`: `linux`, `windows`, or `darwin`.

### `internal/api`

| Client | File | Purpose |
|--------|------|---------|
| `DiscoveryClient` | `supabase.go` | Query `vpn_servers` table for online nodes |
| `AgentClient` | `agent.go` | HTTP calls to per-node agent (`/health`, `/add-peer`, `/replace-peer`) |
| `IPClient` | `ip.go` | Public IP lookup |

### `internal/state`

`Store` persists `VPNState` to JSON with atomic writes (write temp → rename).

### `internal/config`

`Config` struct + `Load()` via Viper. Defaults and validation.

### `internal/ui`

Interactive UX: main menu, server selection, onboarding, confirmations, spinners, tables.

### `pkg/models`

Shared types: `Server`, `PeerRequest`, `PeerResponse`, `WireGuardConfig`, `VPNState`, `EasyVPNError`.

### `pkg/utils`

- `logger.go` — Zap logger with secret redaction
- `validator.go` — IP, URL, port, and required field validation

## Connection Lifecycle

```
User runs "connect"
        │
        ▼
┌───────────────────┐
│ Load existing     │──▶ Already connected? Prompt for key rotation
│ state             │
└─────────┬─────────┘
          ▼
┌───────────────────┐
│ Agent health      │──▶ GET /health (X-API-TOKEN)
│ check             │
└─────────┬─────────┘
          ▼
┌───────────────────┐
│ Generate keys     │──▶ wg genkey | wg pubkey
└─────────┬─────────┘
          ▼
┌───────────────────┐
│ Provision peer    │──▶ POST /add-peer or /replace-peer
└─────────┬─────────┘
          ▼
┌───────────────────┐
│ Build WG config   │──▶ Interface + Peer from PeerResponse
└─────────┬─────────┘
          ▼
┌───────────────────┐
│ Platform adapter  │──▶ wg-quick / wireguard.exe / export .conf
│ CreateTunnel()    │
└─────────┬─────────┘
          ▼
┌───────────────────┐
│ Save state.json   │
└───────────────────┘
```

## Key Design Decisions

### Truth-First Reconciliation

The CLI does not blindly trust `state.json`. On `status`, the reconciler queries the OS and handles:

- **Ghost connection** — State says connected, tunnel is gone → clear state, disable kill-switch
- **Untracked connection** — Tunnel active but state says disconnected → report as "Unknown (Manual)"
- **In sync** — Return persisted state as-is

### Key Rotation on Reconnect

If already connected, `Connect()` warns the user that proceeding will overwrite config and rotate keys. On confirmation, it calls `ReplacePeer` with the old and new public keys instead of `AddPeer`.

### Platform Abstraction

Tunnel management differs radically per OS. The adapter pattern isolates OS-specific shell commands (`wg-quick`, `wireguard.exe`, file export) from business logic.

### Structured Errors

All user-facing failures use `*models.EasyVPNError` with:

- `Code` — Machine-readable identifier
- `Message` — Human explanation
- `Remediation` — Actionable fix
- `Internal` — Wrapped underlying error (for logs)

## External Dependencies

| Service | Protocol | Auth |
|---------|----------|------|
| Supabase PostgREST | HTTPS | Anon key |
| Node Agent | HTTP | `X-API-TOKEN` header |
| IP check service | HTTPS | None |

## Data Models

See [API Integration](api-integration.md) for request/response schemas and Supabase table structure.

## Related

- [State & Reconciliation](state-and-reconciliation.md)
- [Platform Support](platform-support.md)
- [Development](development.md)
