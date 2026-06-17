# Troubleshooting

Common issues and how to resolve them when using EasyVPN CLI.

## Configuration Issues

### "Configuration Missing (.env not found)"

**Symptom:** Onboarding prompt on startup, then exit.

**Fix:**

```bash
cp .env.example .env
# Fill in EASYVPN_API_TOKEN, EASYVPN_SUPABASE_URL, EASYVPN_SUPABASE_ANON_KEY
```

Ensure `.env` is in the project root (or set environment variables directly).

### `[ERR_CONFIG_MISSING] Missing required field`

**Symptom:** One or more required variables are empty.

**Fix:** Verify all three required variables are set and non-empty. No trailing spaces.

### `[ERR_INVALID_INPUT] Invalid Supabase URL`

**Symptom:** URL validation fails on startup.

**Fix:** URL must include scheme and host, e.g. `https://abcdefgh.supabase.co` (not just the project ID).

---

## Discovery Issues

### `[ERR_DISCOVERY_FAILED] Failed to fetch servers`

**Symptom:** `connect` fails after "Discovery failed" spinner.

**Possible causes:**

1. Invalid Supabase URL or anon key
2. No network connectivity
3. `vpn_servers` table empty or no rows with `status = 'online'`
4. Supabase RLS policies blocking anonymous reads

**Fix:**

- Verify credentials in `.env`
- Test Supabase REST API manually
- Confirm servers exist with `status = 'online'` in the database
- Check Row Level Security policies allow `SELECT` for anon role

### "No servers available"

**Symptom:** Server selection fails immediately.

**Fix:** Same as discovery failure. Ensure at least one online server exists in `vpn_servers`.

---

## Agent / Authentication Issues

### `[ERR_AGENT_UNREACHABLE] Node agent is not responding`

**Symptom:** Connect fails at health check.

**Possible causes:**

1. Server is down
2. Agent not running on the node
3. Firewall blocking port (default `5000`)
4. Wrong `public_ip` in database

**Fix:**

- Verify server is online
- Test: `curl -H "X-API-TOKEN: <token>" http://<server-ip>:5000/health`
- Confirm `api_port` in database (defaults to 5000 if missing)

### `[ERR_AUTH_FAILED] Unauthorized access to node agent`

**Symptom:** Health check or add-peer returns 401/403.

**Fix:** Verify `EASYVPN_API_TOKEN` matches the token configured on the node agent.

### `[ERR_INTERNAL] Node agent failed to provision peer`

**Symptom:** Add-peer returns non-auth HTTP error.

**Possible causes:**

1. Server out of internal IP addresses
2. Agent misconfiguration
3. Invalid public key format

**Fix:** Check agent logs on the server. Retry connect to generate fresh keys.

---

## Tunnel Issues

### `[ERR_TUNNEL_FAILED] wg-quick failed` (Linux)

**Symptom:** Tunnel fails to come up.

**Fix:**

```bash
# Run with sudo
sudo easyvpn connect

# Check wg-quick output manually
sudo wg-quick up ~/.easyvpn/wg0.conf

# Verify tools installed
which wg wg-quick
```

### `[ERR_TUNNEL_FAILED] Failed to start WireGuard service` (Windows)

**Symptom:** `wireguard.exe /installmanager` fails.

**Fix:**

- Run terminal as Administrator
- Confirm WireGuard for Windows is installed
- Check `%USERPROFILE%\.easyvpn\wg0.conf` was written

### macOS config not working

**Symptom:** Exported config fails in WireGuard app.

**Fix:**

1. Confirm `~/Desktop/EasyVPN_macOS.conf` exists
2. Re-import into WireGuard app
3. Ensure `wireguard-tools` was installed for key generation (`brew install wireguard-tools`)

---

## Kill-Switch Issues

### No internet after disconnect (Linux)

**Symptom:** Cannot reach internet even after `easyvpn disconnect`.

**Cause:** iptables OUTPUT rules or policy still blocking traffic.

**Fix:**

```bash
sudo iptables -P OUTPUT ACCEPT
sudo iptables -F OUTPUT
```

Always use `easyvpn disconnect` rather than killing the process mid-connect.

### Kill-switch not supported (macOS)

**Symptom:** Error when kill-switch is triggered.

**Expected behavior:** macOS CLI does not manage kill-switch. Use the WireGuard macOS app's built-in kill-switch in its settings.

---

## State Issues

### `[ERR_INTERNAL] Corrupt state file`

**Symptom:** Status or connect fails on state load.

**Fix:**

```bash
easyvpn disconnect
# or manually:
rm ~/.easyvpn/state.json
```

### Status shows connected but tunnel is down

**Symptom:** Ghost connection.

**Expected behavior:** `easyvpn status` runs reconciliation and should auto-heal, showing `Connected: false`.

If stuck:

```bash
easyvpn disconnect
rm ~/.easyvpn/state.json
```

### Status shows "Unknown (Manual)"

**Symptom:** Tunnel active but not managed by EasyVPN.

**Cause:** A WireGuard tunnel was started outside EasyVPN (manual config, another tool).

**Fix:** This is informational. Use `easyvpn disconnect` only if the tunnel was created by EasyVPN. For manual tunnels, deactivate via your WireGuard client.

---

## Dependency Issues

### `[ERR_MISSING_DEPENDENCY] WireGuard Tools missing`

**Platform fixes:**

| OS | Command |
|----|---------|
| Linux | `sudo apt install wireguard wireguard-tools iptables` |
| macOS | `brew install wireguard-tools` or menu option 1 |
| Windows | Install from [wireguard.com/install](https://www.wireguard.com/install/) |

### `[ERR_PERMISSION_DENIED]` (Linux)

**Fix:** Run with root:

```bash
sudo ./bin/easyvpn connect
```

---

## IP Check Issues

### `[ERR_AGENT_UNREACHABLE] Failed to fetch public IP`

**Symptom:** `easyvpn ip` fails.

**Fix:**

- Check internet connectivity
- Verify `EASYVPN_PUBLIC_IP_CHECK_URL` returns JSON `{"ip": "..."}`
- Try default: `https://api.ipify.org?format=json`

---

## Debug Mode

Enable verbose logging:

```env
EASYVPN_LOG_LEVEL=debug
```

Re-run the failing command. Check stdout for reconciler messages and redacted debug output.

Supabase query failures also print `DEBUG: Supabase Raw Error:` to stdout.

---

## Manual Reset (Nuclear Option)

If all else fails:

```bash
# 1. Disconnect
easyvpn disconnect

# 2. Remove all local state
rm -rf ~/.easyvpn/

# 3. Linux: ensure no stale tunnel
sudo wg-quick down ~/.easyvpn/wg0.conf 2>/dev/null
sudo ip link delete wg0 2>/dev/null

# 4. Linux: reset firewall
sudo iptables -P OUTPUT ACCEPT
sudo iptables -F OUTPUT

# 5. Windows: remove firewall rules
netsh advfirewall firewall delete rule name=EasyVPN-Lock
netsh advfirewall firewall delete rule name=EasyVPN-Allow-Server

# 6. Reconfigure and retry
cp .env.example .env
easyvpn setup
easyvpn connect
```

---

## Getting Help

When reporting issues, include:

1. Operating system and version
2. EasyVPN command that failed
3. Full error output (code + remediation)
4. `EASYVPN_LOG_LEVEL=debug` output (redact tokens)
5. Result of `easyvpn setup`

## Related

- [Error Handling](error-handling.md) — Full error code reference
- [Configuration](configuration.md)
- [Platform Support](platform-support.md)
- [State & Reconciliation](state-and-reconciliation.md)
