# Development

Guide for contributing to and developing EasyVPN CLI locally.

## Prerequisites

- Go 1.26.3+
- [Task](https://taskfile.dev/) (optional but recommended)
- Platform WireGuard tools for integration testing
- Valid `.env` with Supabase and API credentials

## Project Structure

```
EasyVPN-CLI/
├── cmd/easyvpn/main.go       # Entry point, Cobra commands, menu loop
├── internal/
│   ├── api/                  # External service clients
│   │   ├── agent.go          # Node agent HTTP client
│   │   ├── ip.go             # Public IP lookup
│   │   └── supabase.go       # Server discovery
│   ├── config/config.go      # Viper configuration
│   ├── core/
│   │   ├── vpn.go            # VPN Manager (connect/disconnect)
│   │   └── reconciler.go     # State vs OS reconciliation
│   ├── platform/
│   │   ├── adapter.go        # PlatformAdapter interface
│   │   ├── linux.go
│   │   ├── windows.go
│   │   └── darwin.go
│   ├── state/store.go        # JSON state persistence
│   └── ui/
│       ├── menu.go           # Survey prompts
│       └── spinner.go        # PTerm spinners/tables
├── pkg/
│   ├── models/
│   │   ├── server.go         # Data types
│   │   └── errors.go         # EasyVPNError
│   └── utils/
│       ├── logger.go         # Zap + redaction
│       └── validator.go      # Input validation
├── docs/                     # Documentation
├── Taskfile.yaml             # Build tasks
├── requirements.sh           # Go dependency installer
├── init_project.sh           # Project scaffolding script
└── go.mod
```

## Task Commands

| Task | Command | Description |
|------|---------|-------------|
| Run | `task run -- <args>` | `go run cmd/easyvpn/main.go <args>` |
| Install | `task install` | Run `requirements.sh` |
| Build | `task build` | Output `bin/easyvpn` |
| Test | `task test` | `go test -v ./...` |
| Tidy | `task tidy` | `go fmt ./...` + `go mod tidy` |

## Local Development Workflow

```bash
# 1. Clone and install
git clone <repo>
cd EasyVPN-CLI
task install

# 2. Configure
cp .env.example .env
# Edit credentials

# 3. Develop with hot reload via go run
task run -- status

# 4. Format and test before committing
task tidy
task test
task build
```

## Key Dependencies

| Package | Version | Usage |
|---------|---------|-------|
| `cobra` | v1.10.2 | CLI framework |
| `viper` | v1.21.0 | Config loading |
| `survey/v2` | v2.3.7 | Interactive prompts |
| `pterm` | v0.12.83 | Terminal UI |
| `zap` | v1.28.0 | Logging |
| `supabase-go` | v0.0.4 | Discovery |
| `resty/v2` | v2.17.2 | HTTP client |

## Coding Conventions

### Error Handling

Always use `*models.EasyVPNError` for user-visible failures. Include remediation text. See [Error Handling](error-handling.md).

### Platform Code

New OS-specific behavior belongs in `internal/platform/`. Implement the `PlatformAdapter` interface and register in `GetAdapter()`.

### State Changes

Use `state.Store` for all persistence. Never write `state.json` directly from other packages.

### UI Feedback

Long operations should use spinners:

```go
ui.StartSpinner("Working...")
ui.UpdateSpinner("Next step...")
ui.StopSpinnerSuccess("Done!")
// or ui.StopSpinnerFail("Failed")
```

### Logging

Use the global `utils.Log` after `utils.Initialize()`:

```go
utils.Log.Debugf("detail: %v", value)
utils.Log.Warn("something unexpected")
```

Secrets in log output are automatically redacted.

## Adding a New CLI Command

1. Define `var myCmd = &cobra.Command{...}` in `cmd/easyvpn/main.go`
2. Register in `init()`: `rootCmd.AddCommand(myCmd)`
3. Optionally wire to interactive menu in `runInteractiveMenu()`
4. Document in [CLI Reference](cli-reference.md)

## Adding a New API Endpoint

1. Add request/response types to `pkg/models/server.go` if needed
2. Implement method on `AgentClient` or `DiscoveryClient` in `internal/api/`
3. Call from `Manager` or command handler
4. Document in [API Integration](api-integration.md)

## Testing

```bash
go test -v ./...
```

Run platform-specific tests on the target OS. Tunnel operations require root/admin and live infrastructure.

### Test Areas to Cover

- Config loading and validation
- State store atomic writes
- Reconciler ghost/untracked detection (mock adapter)
- Validator edge cases
- Error code mapping

## Building for Release

```bash
# Local binary
task build

# Cross-compile (examples)
GOOS=linux GOARCH=amd64 go build -o bin/easyvpn-linux-amd64 ./cmd/easyvpn
GOOS=windows GOARCH=amd64 go build -o bin/easyvpn-windows-amd64.exe ./cmd/easyvpn
GOOS=darwin GOARCH=arm64 go build -o bin/easyvpn-darwin-arm64 ./cmd/easyvpn
```

A `.goreleaser.yaml` placeholder exists for future automated releases.

## Initialization Script

`init_project.sh` creates the directory scaffold and placeholder files. Useful when bootstrapping a fresh project from scratch — not needed for normal development on an existing clone.

## Implementation Guide

The repository includes `Implementation_Guide.md` at the root with the original implementation plan and file inventory. For up-to-date architecture documentation, prefer [Architecture](architecture.md).

## Related

- [Architecture](architecture.md)
- [Installation](installation.md)
- [Error Handling](error-handling.md)
