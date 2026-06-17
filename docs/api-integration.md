# API Integration

EasyVPN communicates with two external services: **Supabase** for server discovery and **Node Agents** for WireGuard peer provisioning.

## Overview

```
┌─────────────┐                    ┌──────────────────┐
│ EasyVPN CLI │──── HTTPS ────────▶│ Supabase         │
│             │    (anon key)      │ vpn_servers table│
└──────┬──────┘                    └──────────────────┘
       │
       │ HTTP (X-API-TOKEN)
       ▼
┌──────────────────┐
│ Node Agent       │
│ :5000 (default)  │
│ /health          │
│ /add-peer        │
│ /replace-peer    │
└──────────────────┘
```

---

## Supabase Discovery

**Client:** `internal/api/supabase.go` — `DiscoveryClient`

### Initialization

```go
supabase.NewClient(cfg.SupabaseUrl, cfg.SupabaseKey, nil)
```

Uses `EASYVPN_SUPABASE_URL` and `EASYVPN_SUPABASE_ANON_KEY`.

### Fetch Active Servers

**Method:** `FetchActiveServers(region string)`

**Query:**

- Table: `vpn_servers`
- Filter: `status = 'online'`
- Optional: `region = <value>` when `region` parameter is non-empty
- Limit: 1000 rows

**Returns:** `[]models.Server`

### Server Model

| JSON Field | Go Field | Type | Description |
|------------|----------|------|-------------|
| `id` | `ID` | string | Unique server identifier |
| `name` | `Name` | string | Display name |
| `region` | `Region` | string | Geographic region |
| `public_ip` | `PublicIP` | string | Node public IP address |
| `endpoint_port` | `EndpointPort` | int | WireGuard listen port |
| `wireguard_public_key` | `WireguardPublicKey` | string | Server WG public key |
| `status` | `Status` | string | `online`, `offline`, etc. |
| `last_heartbeat` | `LastHeartbeat` | time | Last health signal |
| `created_at` | `CreatedAt` | time | Record creation time |
| `api_port` | `APIPort` | int | Agent HTTP port (defaults to 5000) |

**Fallback:** If `api_port` is `0` or missing from the database, the client defaults it to `5000`.

### Error Handling

| Condition | Error Code |
|-----------|------------|
| Client init failure | `ERR_DISCOVERY_FAILED` |
| Query failure | `ERR_DISCOVERY_FAILED` |
| JSON parse failure | `ERR_INTERNAL` |

Debug output prints raw Supabase errors to stdout when queries fail.

---

## Node Agent API

**Client:** `internal/api/agent.go` — `AgentClient`

### Connection

```go
NewAgentClient(ip string, port int, token string)
```

- Base URL: `http://<ip>:<port>`
- Timeout: 10 seconds
- Retries: 2 (2 second wait between)
- Headers:
  - `X-API-TOKEN: <EASYVPN_API_TOKEN>`
  - `Content-Type: application/json`

### Endpoints

#### `GET /health`

Verifies the node agent is online and the API token is valid.

**Success:** HTTP 2xx, no body required.

**Errors:**

| Condition | Error Code |
|-----------|------------|
| Connection failure | `ERR_AGENT_UNREACHABLE` |
| HTTP error (e.g., 401) | `ERR_AUTH_FAILED` |

Called during `Manager.Connect()` before key generation.

#### `POST /add-peer`

Provisions a new WireGuard peer on the server.

**Request body:**

```json
{
  "name": "hostname",
  "public_key": "<client-public-key>"
}
```

Go struct: `models.PeerRequest` (`DeviceName` → JSON `name`, `PublicKey` → JSON `public_key`)

**Response body:**

```json
{
  "client_ip": "10.0.0.2/32",
  "server_public_key": "<server-wg-pubkey>",
  "endpoint": "203.0.113.7:51820",
  "dns": "1.1.1.1",
  "allowed_ips": "0.0.0.0/0,::/0"
}
```

Go struct: `models.PeerResponse`

**Errors:**

| HTTP Status | Error Code |
|-------------|------------|
| 401, 403 | `ERR_AUTH_FAILED` |
| Other 4xx/5xx | `ERR_INTERNAL` |
| Network error | `ERR_AGENT_UNREACHABLE` |

#### `POST /replace-peer`

Rotates an existing peer's public key (used when reconnecting while already connected).

**Request body:**

```json
{
  "old_public_key": "<previous-client-pubkey>",
  "public_key": "<new-client-pubkey>"
}
```

Go struct: `models.ReplacePeerRequest`

**Response:** Same `PeerResponse` schema as `/add-peer`.

**Errors:**

| Condition | Error Code |
|-----------|------------|
| Network error | `ERR_AGENT_UNREACHABLE` |
| HTTP error | `ERR_INTERNAL` |

---

## Public IP Client

**Client:** `internal/api/ip.go` — `IPClient`

### `GET <EASYVPN_PUBLIC_IP_CHECK_URL>`

Default URL: `https://api.ipify.org?format=json`

**Expected response:**

```json
{"ip": "203.0.113.42"}
```

**Errors:** `ERR_AGENT_UNREACHABLE` on network or HTTP failure.

---

## Authentication Summary

| Service | Method | Credential |
|---------|--------|------------|
| Supabase | API key in client init | `EASYVPN_SUPABASE_ANON_KEY` |
| Node Agent | HTTP header | `X-API-TOKEN: EASYVPN_API_TOKEN` |
| IP Check | None | — |

## Security Considerations

- Node agents are contacted over **plain HTTP** on the server's public IP. Ensure agents are only exposed on trusted networks or implement TLS at the infrastructure level.
- API tokens and Supabase keys are redacted from logs. See [Configuration](configuration.md#logging-and-secret-redaction).
- WireGuard private keys are generated locally via `wg genkey` and never sent to Supabase.

## Related

- [Configuration](configuration.md)
- [Architecture](architecture.md)
- [Error Handling](error-handling.md)
