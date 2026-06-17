# EasyVPN CLI Documentation

Welcome to the EasyVPN CLI documentation. This guide covers everything from first-time setup to internal architecture.

## Table of Contents

### User Guides

| Document | Description |
|----------|-------------|
| [Getting Started](getting-started.md) | Install, configure, and make your first connection |
| [Installation](installation.md) | System requirements, dependencies, and build instructions |
| [Configuration](configuration.md) | Environment variables, `.env` file, and config directory |
| [CLI Reference](cli-reference.md) | All commands, interactive menu, and usage examples |
| [Platform Support](platform-support.md) | Linux, Windows, and macOS behavior differences |
| [Troubleshooting](troubleshooting.md) | Common problems and how to fix them |

### Technical Reference

| Document | Description |
|----------|-------------|
| [Architecture](architecture.md) | System design, package layout, and connection lifecycle |
| [API Integration](api-integration.md) | Supabase discovery schema and node agent HTTP API |
| [State & Reconciliation](state-and-reconciliation.md) | Persistent state, ghost connections, and kill-switch |
| [Error Handling](error-handling.md) | Error codes, structure, and remediation patterns |

### Contributing

| Document | Description |
|----------|-------------|
| [Development](development.md) | Local setup, Taskfile commands, coding conventions |

## Quick Links

- [Root README](../README.md) — Project overview and quick start
- [.env.example](../.env.example) — Environment variable template
- [Taskfile.yaml](../Taskfile.yaml) — Build and run tasks

## How EasyVPN Works (30-Second Summary)

1. **Discover** — Query Supabase for VPN servers with `status = 'online'`
2. **Select** — Choose a server from the interactive list
3. **Provision** — Generate WireGuard keys locally; register the public key with the node's agent API
4. **Connect** — Apply the WireGuard configuration using the platform adapter
5. **Track** — Persist connection state to `~/.easyvpn/state.json` for status and key rotation

For the full lifecycle diagram and component breakdown, see [Architecture](architecture.md).
