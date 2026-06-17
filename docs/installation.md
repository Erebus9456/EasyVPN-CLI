# Installation

This document covers system requirements, dependency installation, and building the EasyVPN CLI binary.

## System Requirements

### All Platforms

| Requirement | Details |
|-------------|---------|
| Go | 1.26.3 or later (see `go.mod`) |
| Network | Outbound HTTPS to Supabase; HTTP to VPN node agents |
| Credentials | Valid API token and Supabase project access |

### Linux

| Tool | Purpose |
|------|---------|
| `wg` | WireGuard key generation |
| `wg-quick` | Bring tunnel interfaces up/down |
| `iptables` | Kill-switch firewall rules |
| Root/sudo | Required for `wg-quick` and iptables |

**Install on Debian/Ubuntu:**

```bash
sudo apt update
sudo apt install -y wireguard wireguard-tools iptables
```

### Windows

| Tool | Purpose |
|------|---------|
| WireGuard for Windows | Tunnel service management (`wireguard.exe`) |
| Administrator privileges | Required for tunnel and firewall operations |

Download from [wireguard.com/install](https://www.wireguard.com/install/).

### macOS

| Tool | Purpose |
|------|---------|
| `wg` (wireguard-tools) | Key generation only |
| Homebrew | Optional, for automated install via menu |
| WireGuard macOS app | Required to activate exported configs |

**Install wireguard-tools:**

```bash
brew install wireguard-tools
```

Or use EasyVPN menu option **1) Install all requirements**.

## Installing Go Dependencies

### Using Task

```bash
task install
```

### Using requirements.sh

```bash
chmod +x requirements.sh
./requirements.sh
```

The script installs these Go packages:

| Package | Role |
|---------|------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/viper` | Configuration loading |
| `github.com/AlecAivazis/survey/v2` | Interactive prompts |
| `github.com/pterm/pterm` | Spinners, tables, terminal styling |
| `go.uber.org/zap` | Structured logging with secret redaction |
| `github.com/supabase-community/supabase-go` | Server discovery |
| `github.com/go-resty/resty/v2` | HTTP client for agents and IP checks |

## Building the Binary

### Development Build

```bash
task build
```

Output: `bin/easyvpn`

### Manual Build

```bash
go build -o easyvpn ./cmd/easyvpn
```

### Cross-Compilation

```bash
# Linux amd64
GOOS=linux GOARCH=amd64 go build -o easyvpn-linux-amd64 ./cmd/easyvpn

# Windows amd64
GOOS=windows GOARCH=amd64 go build -o easyvpn-windows-amd64.exe ./cmd/easyvpn

# macOS arm64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o easyvpn-darwin-arm64 ./cmd/easyvpn
```

## Running Without Building

```bash
task run                  # Interactive menu
task run -- connect       # Pass CLI args after --
go run cmd/easyvpn/main.go status
```

## Config Directory Setup

On first run, EasyVPN uses `~/.easyvpn/` (configurable via `EASYVPN_CONFIG_DIR`) for:

- `state.json` — Connection state
- `wg0.conf` — WireGuard config (Linux/Windows)

The directory is created automatically when needed. Linux writes configs with mode `0600`; state files use atomic write (temp file + rename).

## Verifying Installation

```bash
# 1. Build
task build

# 2. Check binary
./bin/easyvpn --help

# 3. Verify platform requirements
./bin/easyvpn setup
```

Expected output for a healthy system: `System is ready for EasyVPN!`

## Project Initialization Script

`init_project.sh` scaffolds the directory structure for a fresh clone. It is primarily used during initial project setup and creates placeholder Go files, `.gitignore`, and `.env.example`. You do not need to run this on a normal install.

## Related

- [Getting Started](getting-started.md)
- [Platform Support](platform-support.md)
- [Development](development.md)
