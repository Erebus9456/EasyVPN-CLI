# Configuration

EasyVPN loads configuration from a `.env` file and environment variables using [Viper](https://github.com/spf13/viper). Environment variables always override `.env` values.

## Configuration Sources (Priority)

1. **Environment variables** (highest priority)
2. **`.env` file** in the project root
3. **`.env` file** in the user home directory (fallback path)
4. **Built-in defaults** (lowest priority)

## Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `EASYVPN_API_TOKEN` | Bearer token sent as `X-API-TOKEN` header to node agents | `sk_live_abc123...` |
| `EASYVPN_SUPABASE_URL` | Supabase project REST API URL | `https://abcdefgh.supabase.co` |
| `EASYVPN_SUPABASE_ANON_KEY` | Supabase anonymous (public) API key | `eyJhbGciOiJIUzI1NiIs...` |

If any required variable is missing on startup, EasyVPN triggers the onboarding flow and exits with instructions to create `.env`.

Validation is performed via `config.Config.Validate()`:

- Required fields must be non-empty
- `EASYVPN_SUPABASE_URL` must be a valid URL with scheme and host

## Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `EASYVPN_NODE_API_BASE_URL` | *(empty)* | Reserved for future centralized node API routing |
| `EASYVPN_DEFAULT_REGION` | *(empty)* | Filter servers by region when set |
| `EASYVPN_PUBLIC_IP_CHECK_URL` | `https://api.ipify.org?format=json` | Endpoint for `easyvpn ip` command |
| `EASYVPN_LOG_LEVEL` | `info` | Log verbosity: `debug`, `info`, `warn`, `error` |
| `EASYVPN_ALLOWED_IPS_DEFAULT` | `0.0.0.0/0,::/0` | Default route-all traffic allowed IPs |
| `EASYVPN_DNS_DEFAULT` | `1.1.1.1` | Default DNS for WireGuard interface |
| `EASYVPN_CONFIG_DIR` | `~/.easyvpn` | Directory for state and WireGuard configs |

## Example `.env` File

```env
# --- REQUIRED ---
EASYVPN_API_TOKEN=your-api-token-here
EASYVPN_SUPABASE_URL=https://your-project.supabase.co
EASYVPN_SUPABASE_ANON_KEY=your-anon-key-here

# --- OPTIONAL ---
EASYVPN_DEFAULT_REGION=us-east
EASYVPN_PUBLIC_IP_CHECK_URL=https://api.ipify.org?format=json
EASYVPN_LOG_LEVEL=debug
EASYVPN_CONFIG_DIR=/home/user/.easyvpn
```

Copy from `.env.example` in the project root.

## Config Directory (`EASYVPN_CONFIG_DIR`)

Default location: `~/.easyvpn/`

| File | Purpose | Permissions |
|------|---------|---------------|
| `state.json` | Current VPN connection metadata | `0600` |
| `wg0.conf` | WireGuard tunnel config (Linux/Windows) | `0600` |
| `state.json.tmp` | Temporary file during atomic writes | Transient |

### State File Schema

See [State & Reconciliation](state-and-reconciliation.md) for the full `VPNState` schema.

## Logging and Secret Redaction

The logger (`pkg/utils/logger.go`) automatically redacts values matching:

- `EASYVPN_API_TOKEN`
- `EASYVPN_SUPABASE_ANON_KEY`

These are passed to `utils.Initialize()` at startup. Set `EASYVPN_LOG_LEVEL=debug` for verbose reconciliation and kill-switch logs.

## Onboarding Flow

When `.env` is missing or `EASYVPN_API_TOKEN` is empty:

1. EasyVPN displays a configuration warning
2. User chooses from:
   - **Option 1:** Provide path to existing `.env` *(planned)*
   - **Option 2:** Enter values individually *(planned)*
   - **Option 3:** Print template and exit
3. Application exits until `.env` is created

## Loading Implementation

Configuration is loaded in `internal/config/config.go`:

```go
v.SetConfigName(".env")
v.SetConfigType("env")
v.AddConfigPath(".")   // Project root
v.AddConfigPath(home)  // Home directory
v.AutomaticEnv()
```

Viper's env key replacer maps dots to underscores, so env vars use the `EASYVPN_` prefix directly.

## Related

- [Getting Started](getting-started.md)
- [API Integration](api-integration.md)
- [State & Reconciliation](state-and-reconciliation.md)
