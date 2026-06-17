# EasyVPN CLI

A cross-platform command-line client for connecting to EasyVPN WireGuard servers. EasyVPN discovers VPN nodes from a Supabase registry, provisions peers through per-node agents, and manages tunnels with platform-specific adapters for Linux, Windows, and macOS.

## Features

- **Server discovery** — Fetches online VPN servers from Supabase (`vpn_servers` table)
- **Automated provisioning** — Registers WireGuard peers via node agents with API token auth
- **Key rotation** — Replaces existing peer keys when reconnecting to a different server
- **Cross-platform tunnels** — Native `wg-quick` on Linux, WireGuard service on Windows, config export on macOS
- **Kill-switch** — Firewall-based leak protection on Linux and Windows (enabled by default on connect)
- **State reconciliation** — Detects and heals "ghost" connections when tunnels are closed externally
- **Interactive menu** — TUI-driven workflow with spinners, tables, and guided prompts
- **Structured errors** — Machine-readable error codes with actionable remediation hints

## Quick Start

```bash
# 1. Clone and install Go dependencies
git clone https://github.com/Erebus9456/EasyVPN-CLI.git
cd EasyVPN-CLI
task install   # or: ./requirements.sh

# 2. Configure credentials
cp .env.example .env
# Edit .env with your EASYVPN_API_TOKEN, EASYVPN_SUPABASE_URL, and EASYVPN_SUPABASE_ANON_KEY

# 3. Run
task run                  # Interactive menu
task run -- connect       # Connect to a server
task run -- status        # Check connection status
```

See [Getting Started](docs/getting-started.md) for the full walkthrough.

## Requirements

| Component | Version |
|-----------|---------|
| Go | 1.26+ |
| WireGuard tools | Platform-specific (see [Platform Support](docs/platform-support.md)) |
| Credentials | API token + Supabase URL and anon key |

## CLI Commands

| Command | Description |
|---------|-------------|
| `easyvpn` | Launch the interactive main menu |
| `easyvpn setup` | Verify and install system requirements |
| `easyvpn connect` | Select a server and establish a VPN tunnel |
| `easyvpn disconnect` | Tear down the tunnel and clear state |
| `easyvpn status` | Show current connection state (reconciled with OS) |
| `easyvpn ip` | Display your current public IP address |

Full command reference: [CLI Reference](docs/cli-reference.md)

## Configuration

Required environment variables (via `.env` or shell):

```env
EASYVPN_API_TOKEN=your-api-token
EASYVPN_SUPABASE_URL=https://your-project.supabase.co
EASYVPN_SUPABASE_ANON_KEY=your-anon-key
```

Optional settings include log level, DNS defaults, allowed IPs, and config directory. See [Configuration](docs/configuration.md) for all variables and defaults.

## Platform Behavior

| Platform | Tunnel Management | Kill-Switch |
|----------|-------------------|-------------|
| **Linux** | `wg-quick up/down` via `~/.easyvpn/wg0.conf` | `iptables` rules |
| **Windows** | WireGuard service via `wireguard.exe` | Windows Firewall (`netsh`) |
| **macOS** | Exports `EasyVPN_macOS.conf` to Desktop | Use WireGuard app settings |

Details: [Platform Support](docs/platform-support.md)

## Architecture Overview

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│  CLI / UI   │────▶│  VPN Manager │────▶│ Platform Adapter│
└─────────────┘     └──────┬───────┘     └─────────────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │ Supabase │ │  Agent   │ │  State   │
        │ Discovery│ │  Client  │ │  Store   │
        └──────────┘ └──────────┘ └──────────┘
```

Deep dive: [Architecture](docs/architecture.md)

## Documentation

| Topic | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | First-time setup and first connection |
| [Installation](docs/installation.md) | Dependencies, build, and deployment |
| [Configuration](docs/configuration.md) | Environment variables and config directory |
| [CLI Reference](docs/cli-reference.md) | Commands, flags, and interactive menu |
| [Architecture](docs/architecture.md) | Package layout, data flow, and design decisions |
| [Platform Support](docs/platform-support.md) | Linux, Windows, and macOS specifics |
| [API Integration](docs/api-integration.md) | Supabase discovery and node agent API |
| [State & Reconciliation](docs/state-and-reconciliation.md) | `state.json`, ghost connections, kill-switch |
| [Error Handling](docs/error-handling.md) | Error codes and remediation |
| [Development](docs/development.md) | Local dev workflow, Taskfile, project structure |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and fixes |

## Project Structure

```
EasyVPN-CLI/
├── cmd/easyvpn/          # Application entry point
├── internal/
│   ├── api/              # Supabase, agent, and IP clients
│   ├── config/           # Viper-based configuration loading
│   ├── core/             # VPN manager and state reconciler
│   ├── platform/         # OS-specific tunnel adapters
│   ├── state/            # Persistent state store
│   └── ui/               # Interactive menus and spinners
├── pkg/
│   ├── models/           # Shared data types and errors
│   └── utils/            # Logger and validators
├── docs/                 # Documentation (you are here)
├── Taskfile.yaml         # Task runner definitions
└── .env.example          # Environment variable template
```

## Development

```bash
task install    # Install Go dependencies
task build      # Build binary to bin/easyvpn
task test       # Run tests
task tidy       # Format and tidy go.mod
```

See [Development Guide](docs/development.md) for details.

## License

See repository license file for terms.
