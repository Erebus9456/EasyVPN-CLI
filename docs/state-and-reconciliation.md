# State & Reconciliation

EasyVPN tracks VPN connection state on disk and reconciles it with the operating system's actual tunnel status.

## State Store

**Package:** `internal/state`  
**File:** `~/.easyvpn/state.json` (configurable via `EASYVPN_CONFIG_DIR`)

### VPNState Schema

```json
{
  "is_connected": true,
  "server_id": "uuid-of-server",
  "server_name": "US-East-1",
  "server_public_ip": "203.0.113.7",
  "client_public_key": "abc123...",
  "interface_name": "wg0",
  "connected_at": "2026-06-17T10:30:00Z",
  "kill_switch_enabled": true
}
```

| Field | Type | Purpose |
|-------|------|---------|
| `is_connected` | bool | Whether CLI believes a tunnel is active |
| `server_id` | string | Supabase server ID |
| `server_name` | string | Human-readable server name |
| `server_public_ip` | string | Used for kill-switch allow rules |
| `client_public_key` | string | Stored for key rotation via `/replace-peer` |
| `interface_name` | string | WireGuard interface (default `wg0`) |
| `connected_at` | timestamp | When connection was established |
| `kill_switch_enabled` | bool | Whether kill-switch was enabled on connect |

### Persistence Operations

| Method | Behavior |
|--------|----------|
| `Save(state)` | Atomic write: `state.json.tmp` → rename to `state.json` |
| `Load()` | Read and parse JSON; returns disconnected state if file missing |
| `Clear()` | Save `{is_connected: false}` |

**File permissions:** `0600` (owner read/write only)

**Corrupt state:** `Load()` returns `ERR_INTERNAL` with remediation: "Run 'easyvpn disconnect' to reset"

---

## Reconciler

**Package:** `internal/core/reconciler.go`

The reconciler implements a **truth-first** model: OS reality takes precedence over persisted state when they disagree.

### Sync Flow

```
Load state.json
       │
       ▼
Query OS via adapter.GetStatus()
       │
       ├── OS query fails ──▶ Return persisted state (unverified)
       │
       ▼
Compare persisted vs real
       │
       ├── Ghost connection ──▶ Clear state, disable kill-switch
       ├── Untracked connection ──▶ Report manual/unknown tunnel
       └── In sync ──▶ Return persisted state
```

### Ghost Connection

**Condition:** `state.is_connected == true` but OS reports tunnel inactive.

**Cause:** User closed tunnel externally (e.g., `wg-quick down`, WireGuard app, system reboot).

**Healing actions:**

1. Log warning: `Detected 'Ghost Connection'`
2. Disable kill-switch (prevents network lockout)
3. Clear `state.json`
4. Return `{is_connected: false}`

### Untracked Connection

**Condition:** `state.is_connected == false` but OS reports active tunnel.

**Cause:** Manual WireGuard config activated outside EasyVPN.

**Behavior:** Does **not** adopt the tunnel into state (avoids corrupting manual configs). Returns:

```go
VPNState{
    IsConnected:   true,
    InterfaceName: realStatus.Interface,
    ServerName:    "Unknown (Manual)",
}
```

### In Sync

Both persisted state and OS agree. Return persisted state unchanged.

---

## Kill-Switch

Kill-switch is enabled by default when connecting (`Manager.Connect(..., useKillSwitch=true)`).

### State Tracking

`kill_switch_enabled` is saved in `state.json` and displayed in `easyvpn status`.

### Disconnect Behavior

`Manager.Disconnect()` **always** disables kill-switch first, before tearing down the tunnel. This prevents leaving the user blocked from the internet.

### Reconciler Kill-Switch Check

`Reconciler.CheckKillSwitch()` re-applies firewall rules if kill-switch is enabled and connected. Currently logs at debug level; full re-application is stubbed for future hardening.

### Platform Implementation

| Platform | Mechanism | See |
|----------|-----------|-----|
| Linux | `iptables` OUTPUT chain | [Platform Support](platform-support.md#kill-switch-iptables) |
| Windows | `netsh advfirewall` rules | [Platform Support](platform-support.md#kill-switch-netsh) |
| macOS | Not supported via CLI | Use WireGuard app |

---

## Key Rotation State

When reconnecting while already connected:

1. `Load()` returns existing state with `client_public_key`
2. User confirms overwrite
3. `ReplacePeer(oldPubKey, newPubKey)` is called
4. New `client_public_key` is saved in state

The stored public key is essential for server-side peer replacement without orphaning the old key.

---

## State File Locations by Platform

| Platform | State | WireGuard Config |
|----------|-------|------------------|
| Linux | `~/.easyvpn/state.json` | `~/.easyvpn/wg0.conf` |
| Windows | `%USERPROFILE%\.easyvpn\state.json` | `%USERPROFILE%\.easyvpn\wg0.conf` |
| macOS | `~/.easyvpn/state.json` | `~/Desktop/EasyVPN_macOS.conf` (export only) |

---

## Manual Recovery

If state becomes inconsistent:

```bash
# Reset state and tear down tunnel
easyvpn disconnect

# If disconnect fails, manually remove state
rm ~/.easyvpn/state.json

# Linux: bring down tunnel manually
sudo wg-quick down ~/.easyvpn/wg0.conf

# Linux: restore iptables if kill-switch left rules
sudo iptables -P OUTPUT ACCEPT
sudo iptables -F OUTPUT
```

See [Troubleshooting](troubleshooting.md) for more recovery steps.

## Related

- [Architecture](architecture.md)
- [Platform Support](platform-support.md)
- [CLI Reference](cli-reference.md#easyvpn-status)
