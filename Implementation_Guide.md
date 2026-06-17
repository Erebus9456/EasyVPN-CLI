# 🚀 EasyVPN Implementation Plan

---

# 📁 File Inventory

## Root
```text
Taskfile.yaml
.goreleaser.yaml
```

## Command Entry Point
```text
cmd/
└── easyvpn/
    └── main.go
```

## Internal Packages

### API Layer
```text
internal/api/
├── supabase.go
└── agent.go
```

### Configuration
```text
internal/config/
└── config.go
```

### Core Business Logic
```text
internal/core/
├── reconciler.go
└── vpn.go
```

### Platform Implementations
```text
internal/platform/
├── adapter.go
├── linux.go
├── windows.go
└── darwin.go
```

### State Management
```text
internal/state/
└── store.go
```

### User Interface
```text
internal/ui/
├── menu.go
└── spinner.go
```

## Shared Packages

### Models
```text
pkg/models/
├── server.go
└── errors.go
```

### Utilities
```text
pkg/utils/
├── logger.go
└── validator.go
```

---

# 🛣️ Implementation Roadmap

## Phase 1 — The Foundations
### *The "Beast-Mode" Contract*

Everything else depends on reliable error handling, validation, and logging.

| File | Purpose |
|--------|---------|
| `pkg/models/errors.go` | Define custom error structures and remediation logic |
| `pkg/utils/logger.go` | Configure structured logging with secret redaction (Zap) |
| `pkg/utils/validator.go` | Validate IP addresses, URLs, public keys, and configuration values |
| `pkg/models/server.go` | Define Server and VPN Profile data models |

### Deliverables
- Centralized error handling
- Structured logging
- Input validation framework
- Core domain models

---

## Phase 2 — Configuration & State
### *The "Brain"*

Before connecting to anything, the application must understand its configuration and current state.

| Step | File | Purpose |
|--------|---------|---------|
| 5 | `internal/config/config.go` | Load `.env` and `config.json` using Viper |
| 6 | `internal/state/store.go` | Read/write `state.json` using the Truth-First engine |

### Deliverables
- Configuration loading
- Environment variable support
- Persistent state management
- Recovery after restarts

---

## Phase 3 — CLI Entry & User Interface
### *The "Face"*

Build the user-facing experience and establish application entry points.

| Step | File | Purpose |
|--------|---------|---------|
| 7 | `cmd/easyvpn/main.go` | Configure Cobra commands (`connect`, `disconnect`, `status`) |
| 8 | `internal/ui/menu.go` | Interactive numeric selection menus |
| 9 | `internal/ui/spinner.go` | Progress indicators for long-running operations |

### Deliverables
- Functional CLI
- Interactive menus
- User feedback system

---

## Phase 4 — Remote Communications
### *The "Comms"*

Enable communication with external services.

| Step | File | Purpose |
|--------|---------|---------|
| 10 | `internal/api/supabase.go` | Fetch server inventory from Supabase |
| 11 | `internal/api/agent.go` | Communicate with the Node Agent (`/add-peer`) |

### Deliverables
- Server discovery
- Peer provisioning
- External API integration

---

## Phase 5 — OS Interaction
### *The "Muscle"*

The most platform-specific layer: interacting with the host operating system and networking stack.

| Step | File | Purpose |
|--------|---------|---------|
| 12 | `internal/platform/adapter.go` | Define the platform abstraction interface |
| 13 | `internal/platform/linux.go` | WireGuard integration for Linux (`wg-quick`, `iptables`) |
| 14 | `internal/platform/windows.go` | WireGuard integration for Windows |
| 15 | `internal/platform/darwin.go` | Configuration export workflow for macOS |

### Deliverables
- Cross-platform abstraction
- Linux networking support
- Windows support
- macOS support

---

## Phase 6 — Logic Orchestration
### *The "Core"*

The orchestration layer that binds the UI, APIs, state management, and platform implementations together.

| Step | File | Purpose |
|--------|---------|---------|
| 16 | `internal/core/vpn.go` | High-level VPN workflow orchestration |
| 17 | `internal/core/reconciler.go` | Self-healing reconciliation engine |

### Example Workflow

```text
Connect
 ├─ Discover Server
 ├─ Provision Peer
 ├─ Generate Config
 ├─ Apply Platform Changes
 ├─ Persist State
 └─ Verify Connectivity
```

### Deliverables
- End-to-end connection workflow
- State reconciliation
- Self-healing behavior
- Platform-independent VPN operations

---

# 🎯 Build Order Summary

```text
1.  errors.go
2.  logger.go
3.  validator.go
4.  server.go

5.  config.go
6.  store.go

7.  main.go
8.  menu.go
9.  spinner.go

10. supabase.go
11. agent.go

12. adapter.go
13. linux.go
14. windows.go
15. darwin.go

16. vpn.go
17. reconciler.go
```

---

# 🏁 End State

Once all phases are complete, EasyVPN will provide:

✅ Cross-platform VPN provisioning  
✅ Automatic peer management  
✅ Self-healing state reconciliation  
✅ Interactive CLI experience  
✅ Secure configuration handling  
✅ Structured observability and logging  
✅ Production-ready release pipeline