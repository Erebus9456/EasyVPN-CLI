# Platform Support

EasyVPN uses platform-specific adapters to manage WireGuard tunnels. Behavior differs significantly across Linux, Windows, and macOS.

## Adapter Selection

At startup, `platform.GetAdapter()` inspects `runtime.GOOS`:

| GOOS | Adapter | File |
|------|---------|------|
| `linux` | `LinuxAdapter` | `internal/platform/linux.go` |
| `windows` | `WindowsAdapter` | `internal/platform/windows.go` |
| `darwin` | `DarwinAdapter` | `internal/platform/darwin.go` |

Unsupported OS returns `ERR_INTERNAL: Unsupported Operating System`.

## Comparison Matrix

| Feature | Linux | Windows | macOS |
|---------|-------|---------|-------|
| Tunnel management | `wg-quick` | WireGuard service | Config export only |
| Privilege required | Root (`euid == 0`) | Administrator | None (export mode) |
| Config location | `~/.easyvpn/wg0.conf` | `%USERPROFILE%\.easyvpn\wg0.conf` | `~/Desktop/EasyVPN_macOS.conf` |
| Kill-switch | `iptables` | `netsh` advfirewall | Not supported via CLI |
| Status detection | `wg show wg0 dump` | `sc query WireGuardTunnel$wg0` | Always reports inactive |
| Auto-install deps | Manual guidance | Manual guidance | Homebrew (`brew install wireguard-tools`) |
| Interactive menu | Full VPN menu | Full VPN menu | Export-focused menu |

---

## Linux

### Requirements

- `wg` — Key generation
- `wg-quick` — Interface lifecycle
- `iptables` — Kill-switch
- Root privileges for tunnel and firewall operations

### Creating a Tunnel

1. Writes WireGuard config to `~/.easyvpn/wg0.conf` (mode `0600`)
2. Runs `wg-quick up ~/.easyvpn/wg0.conf`
3. Interface name: `wg0`

### Destroying a Tunnel

Runs `wg-quick down ~/.easyvpn/wg0.conf`. Silently succeeds if config file does not exist.

### Kill-Switch (`iptables`)

When enabled:

```bash
iptables -A OUTPUT -d <serverIP> -j ACCEPT    # Allow VPN server
iptables -A OUTPUT -o wg0 -j ACCEPT           # Allow tunnel traffic
iptables -P OUTPUT DROP                        # Block everything else
```

When disabled:

```bash
iptables -P OUTPUT ACCEPT
iptables -F OUTPUT
```

> **Note:** Kill-switch rules are appended without tracking rule IDs. Repeated enable/disable cycles may accumulate rules. Run disconnect to restore defaults.

### Status Detection

Parses `wg show wg0 dump`. Returns `Active: true` with interface `wg0` if output is non-empty.

### Install Dependencies

Automatic install is not supported. User is directed to:

```bash
sudo apt update && sudo apt install -y wireguard wireguard-tools iptables
```

---

## Windows

### Requirements

- WireGuard for Windows (`wireguard.exe` in PATH)
- Administrator privileges

### Privilege Check

Attempts to open `\\.\PHYSICALDRIVE0`. Success indicates admin rights.

### Creating a Tunnel

1. Writes config to `%USERPROFILE%\.easyvpn\wg0.conf`
2. Runs `wireguard.exe /installmanager <configPath>`
3. Registers and starts WireGuard tunnel service

### Destroying a Tunnel

Runs `wireguard.exe /uninstallmanager wg0`. Ignores errors if tunnel does not exist.

### Kill-Switch (`netsh`)

When enabled:

```bash
netsh advfirewall firewall add rule name=EasyVPN-Lock dir=out action=block
netsh advfirewall firewall add rule name=EasyVPN-Allow-Server dir=out action=allow remoteip=<serverIP>
```

When disabled, deletes both rules by name.

### Status Detection

Queries Windows service: `sc query WireGuardTunnel$wg0`. Active if output contains `RUNNING`.

### Install Dependencies

Manual install required from [wireguard.com/install](https://www.wireguard.com/install/).

---

## macOS (Darwin)

macOS does **not** manage tunnels from the CLI. EasyVPN operates in **export mode**.

### Requirements

- `wg` binary (from `wireguard-tools` via Homebrew)
- Official WireGuard macOS app for activation

### Creating a Tunnel (Export)

1. Builds standard WireGuard `.conf` content
2. Writes to `~/Desktop/EasyVPN_macOS.conf` (mode `0600`)
3. Prints step-by-step import instructions

**User steps after export:**

1. Open the WireGuard app
2. Import tunnel from file
3. Select `EasyVPN_macOS.conf` from Desktop
4. Click Activate

### Destroy Tunnel

No-op — tunnel is managed by the WireGuard app, not the CLI.

### Kill-Switch

Returns error directing users to the WireGuard app's built-in kill-switch settings:

```
Kill-switch via CLI is not supported on macOS
Please use the 'Kill-switch' feature inside the official WireGuard macOS app settings.
```

### Status Detection

Always returns `Active: false` since the CLI does not manage the tunnel.

### Install Dependencies

If Homebrew is installed:

```bash
brew install wireguard-tools
```

Triggered from menu option 1 or `DarwinAdapter.InstallDependencies()`.

### Interactive Menu

macOS shows a simplified menu (install, IP check, export config) without disconnect/status/kill-switch options.

---

## WireGuard Config Format

All platforms generate configs in this structure:

```ini
[Interface]
PrivateKey = <client-private-key>
Address = <assigned-client-ip>
DNS = <dns-servers>

[Peer]
PublicKey = <server-public-key>
Endpoint = <server-ip:port>
AllowedIPs = <routes>
PersistentKeepalive = 25
```

Values come from the node agent's `PeerResponse` after provisioning.

## Related

- [Installation](installation.md)
- [CLI Reference](cli-reference.md)
- [Architecture](architecture.md)
