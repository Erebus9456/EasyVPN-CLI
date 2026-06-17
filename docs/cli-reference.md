# CLI Reference

Complete reference for EasyVPN CLI commands, the interactive menu, and usage patterns.

## Global Behavior

```bash
easyvpn [command]
```

Running `easyvpn` with no subcommand launches the **interactive main menu**.

On startup, EasyVPN:

1. Loads configuration from `.env` / environment
2. Triggers onboarding if credentials are missing
3. Initializes structured logging with secret redaction
4. Creates the VPN `Manager` with platform adapter

Exit codes:

- `0` — Success or graceful menu exit
- `1` — Error (prints remediation hint for `EasyVPNError`)

## Commands

### `easyvpn` (default)

Launches the interactive menu loop. After each action, prompts **Press Enter to return to menu**.

### `easyvpn setup`

Verify system requirements for the current platform.

```bash
easyvpn setup
```

**Behavior:**

- Calls `PlatformAdapter.CheckRequirements()`
- Reports each missing dependency with remediation text
- On success: `System is ready for EasyVPN!`

**Platform checks:**

| OS | Dependencies verified |
|----|----------------------|
| Linux | `wg`, `wg-quick`, `iptables` |
| Windows | `wireguard.exe` in PATH |
| macOS | `wg` (wireguard-tools) |

### `easyvpn connect`

Select a server and establish a VPN connection.

```bash
easyvpn connect
```

**Flow:**

1. Fetch active servers from Supabase (`status = 'online'`)
2. Display server count spinner
3. Interactive server selection (`[region] name (ip)`)
4. Call `Manager.Connect(server, hostname, killSwitch=true)`

**Kill-switch:** Enabled by default on connect (`useKillSwitch=true`).

**Key rotation:** If already connected, prompts for confirmation before overwriting config and rotating keys.

### `easyvpn disconnect`

Tear down the active VPN tunnel.

```bash
easyvpn disconnect
```

**Flow:**

1. Disable kill-switch (prevents lockout)
2. Destroy tunnel via platform adapter
3. Clear `state.json`

### `easyvpn status`

Show current VPN connection state, reconciled with OS reality.

```bash
easyvpn status
```

**Output table:**

| Property | Description |
|----------|-------------|
| Connected | `true` / `false` |
| Server | Server name or `Unknown (Manual)` |
| Interface | WireGuard interface name (e.g., `wg0`) |
| Kill-Switch | Whether kill-switch is enabled in state |

Uses `Reconciler.Sync()` to detect ghost or untracked connections before display.

### `easyvpn ip`

Query and display your current public IP address.

```bash
easyvpn ip
```

Uses `EASYVPN_PUBLIC_IP_CHECK_URL` (default: `https://api.ipify.org?format=json`). Expects JSON response with `{"ip": "x.x.x.x"}`.

## Interactive Menu

The menu adapts to the operating system.

### macOS Menu

```
EasyVPN Main Menu:
  1) Install all requirements (wireguard-tools)
  2) Check Current Public IP
  3) Export macOS WireGuard Config
  0) Exit
```

| Option | Maps to |
|--------|---------|
| 1 | `easyvpn setup` (+ Homebrew install on macOS) |
| 2 | `easyvpn ip` |
| 3 | `easyvpn connect` (exports config to Desktop) |
| 0 | Exit |

### Linux / Windows Menu

```
EasyVPN Main Menu:
  1) Install all requirements and prepare machine
  2) List VPN Servers
  3) Connect to VPN
  4) Disconnect VPN
  5) Show Connection Status
  6) Kill-switch Management
  7) Check Current Public IP
  8) Export macOS WireGuard Config
  0) Exit
```

| Option | Maps to |
|--------|---------|
| 1 | `easyvpn setup` |
| 2 | `easyvpn connect` (server list is shown during connect) |
| 3 | `easyvpn connect` |
| 4 | `easyvpn disconnect` |
| 5 | `easyvpn status` |
| 6 | Not implemented — prints "Feature not available on this platform." |
| 7 | `easyvpn ip` |
| 8 | `easyvpn connect` (no-op on non-macOS for export path) |
| 0 | Exit |

## Task Runner Shortcuts

From the project root with [Task](https://taskfile.dev/):

```bash
task run                  # Interactive menu
task run -- connect       # Connect command
task run -- disconnect    # Disconnect command
task run -- status        # Status command
task run -- ip            # IP command
task run -- setup         # Setup command
```

## Error Output Format

When an `EasyVPNError` occurs:

```
❌ [ERR_AUTH_FAILED] Unauthorized access to node agent
💡 Fix: Your EASYVPN_API_TOKEN is likely invalid for this server
```

Generic errors:

```
❌ Error: connection refused
```

See [Error Handling](error-handling.md) for all error codes.

## UI Components

| Component | Library | Usage |
|-----------|---------|-------|
| Menus / prompts | `survey/v2` | Server selection, confirmations, onboarding |
| Spinners | `pterm` | Progress during connect, status, IP check |
| Tables | `pterm` | Status output |

Spinner states: start → update text → success (✓) or fail (✗).

## Related

- [Getting Started](getting-started.md)
- [Platform Support](platform-support.md)
- [Troubleshooting](troubleshooting.md)
