# Error Handling

EasyVPN uses a structured error type with machine-readable codes, human messages, and actionable remediation hints.

## Error Structure

**Type:** `*models.EasyVPNError` (`pkg/models/errors.go`)

```go
type EasyVPNError struct {
    Code        ErrorCode  // Machine-readable identifier
    Message     string     // Human-readable explanation
    Remediation string     // Steps to fix the issue
    Internal    error      // Underlying error (for logging)
}
```

### Creating Errors

```go
models.NewError(code, message, remediation, internalErr)
models.WrapError(code, message, remediation, internalErr)
```

### Display Format

The CLI handler (`handleError` in `main.go`) prints:

```
❌ [ERR_AUTH_FAILED] Unauthorized access to node agent
💡 Fix: Your EASYVPN_API_TOKEN is likely invalid for this server
```

Non-`EasyVPNError` values print as generic errors and exit with code 1.

---

## Error Codes

### System Errors

| Code | Constant | Typical Cause | Remediation |
|------|----------|---------------|-------------|
| `ERR_PERMISSION_DENIED` | `ErrPermissionDenied` | Missing root/admin | Run with elevated privileges |
| `ERR_MISSING_DEPENDENCY` | `ErrMissingDependency` | `wg`, `wg-quick`, etc. missing | Install platform dependencies |
| `ERR_INTERNAL` | `ErrInternal` | Unexpected failure | Varies by context |

### Configuration Errors

| Code | Constant | Typical Cause | Remediation |
|------|----------|---------------|-------------|
| `ERR_CONFIG_MISSING` | `ErrConfigMissing` | Empty required env var | Set variable in `.env` |
| `ERR_INVALID_INPUT` | `ErrInvalidInput` | Malformed URL, IP, or port | Correct the value in config |

### Network / API Errors

| Code | Constant | Typical Cause | Remediation |
|------|----------|---------------|-------------|
| `ERR_AUTH_FAILED` | `ErrAuthFailed` | Invalid API token | Check `EASYVPN_API_TOKEN` |
| `ERR_DISCOVERY_FAILED` | `ErrDiscoveryFailed` | Supabase unreachable or empty | Check URL, key, network |
| `ERR_AGENT_UNREACHABLE` | `ErrAgentUnreachable` | Node agent down or blocked | Check server IP, port, firewall |

### VPN Errors

| Code | Constant | Typical Cause | Remediation |
|------|----------|---------------|-------------|
| `ERR_TUNNEL_FAILED` | `ErrTunnelFailed` | `wg-quick` or WireGuard service failed | Check privileges, config, logs |
| `ERR_KILLSWITCH_FAILED` | `ErrKillSwitchFail` | Firewall rule application failed | Check iptables/netsh permissions |

---

## Errors by Component

### Configuration (`internal/config`)

| Scenario | Code |
|----------|------|
| `.env` read error (not "not found") | `ERR_INTERNAL` |
| Missing required field | `ERR_CONFIG_MISSING` |
| Invalid Supabase URL | `ERR_INVALID_INPUT` |
| Config directory creation failed | `ERR_INTERNAL` |

### Discovery (`internal/api/supabase.go`)

| Scenario | Code |
|----------|------|
| Supabase client init failed | `ERR_DISCOVERY_FAILED` |
| Query execution failed | `ERR_DISCOVERY_FAILED` |
| JSON unmarshal failed | `ERR_INTERNAL` |
| No servers in list (UI) | `ERR_DISCOVERY_FAILED` |

### Agent (`internal/api/agent.go`)

| Scenario | Code |
|----------|------|
| Health check connection error | `ERR_AGENT_UNREACHABLE` |
| Health check HTTP error | `ERR_AUTH_FAILED` |
| Add peer connection error | `ERR_AGENT_UNREACHABLE` |
| Add peer 401/403 | `ERR_AUTH_FAILED` |
| Add peer other HTTP error | `ERR_INTERNAL` |
| Replace peer errors | `ERR_AGENT_UNREACHABLE` / `ERR_INTERNAL` |

### VPN Manager (`internal/core/vpn.go`)

| Scenario | Code |
|----------|------|
| User aborts reconnect confirmation | `ERR_INTERNAL` |
| Key generation failed | `ERR_INTERNAL` |
| Platform adapter not initialized | `ERR_INTERNAL` |

### State Store (`internal/state/store.go`)

| Scenario | Code |
|----------|------|
| JSON marshal/unmarshal failed | `ERR_INTERNAL` |
| File read/write failed | `ERR_INTERNAL` |
| Corrupt state file | `ERR_INTERNAL` |

### Platform Adapters

| Platform | Scenario | Code |
|----------|----------|------|
| All | Unsupported OS | `ERR_INTERNAL` |
| Linux | `wg-quick` failure | `ERR_TUNNEL_FAILED` |
| Linux | Auto-install not supported | `ERR_PERMISSION_DENIED` |
| Windows | WireGuard not found | `ERR_MISSING_DEPENDENCY` |
| Windows | Service install failed | `ERR_TUNNEL_FAILED` |
| macOS | `wg` missing | `ERR_MISSING_DEPENDENCY` |
| macOS | Homebrew missing | `ERR_MISSING_DEPENDENCY` |
| macOS | Kill-switch requested | `ERR_INTERNAL` |
| macOS | Export write failed | `ERR_INTERNAL` |

### Validation (`pkg/utils/validator.go`)

| Scenario | Code |
|----------|------|
| Invalid IP/CIDR | `ERR_INVALID_INPUT` |
| Invalid URL | `ERR_INVALID_INPUT` |
| Invalid port | `ERR_INVALID_INPUT` |
| Empty required field | `ERR_CONFIG_MISSING` |

---

## Logging vs User Errors

- **User-facing:** Printed via `handleError()` with code, message, and remediation
- **Internal:** `Internal` field and `utils.Log` debug/warn output (secrets redacted)

Enable debug logging:

```env
EASYVPN_LOG_LEVEL=debug
```

Reconciler ghost connection detection logs at `warn` level.

---

## Best Practices for Contributors

When adding new failure paths:

1. Use `models.NewError()` with a specific `ErrorCode`
2. Write a clear `Message` explaining what happened
3. Always provide actionable `Remediation` text
4. Wrap the underlying `error` in `Internal` for debugging
5. Return `*EasyVPNError` from functions that may fail user-visible operations

## Related

- [Troubleshooting](troubleshooting.md)
- [API Integration](api-integration.md)
- [Configuration](configuration.md)
