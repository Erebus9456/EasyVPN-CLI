# Getting Started

This guide walks you through installing EasyVPN CLI, configuring credentials, and making your first VPN connection.

## Prerequisites

Before you begin, ensure you have:

1. **Go 1.26 or later** installed ([golang.org/dl](https://go.dev/dl/))
2. **EasyVPN credentials**:
   - `EASYVPN_API_TOKEN` — Shared secret for node agent authentication
   - `EASYVPN_SUPABASE_URL` — Your Supabase project URL
   - `EASYVPN_SUPABASE_ANON_KEY` — Supabase anonymous/public API key
3. **WireGuard tools** for your platform (see [Platform Support](platform-support.md))

## Step 1: Clone the Repository

```bash
git clone https://github.com/Erebus9456/EasyVPN-CLI.git
cd EasyVPN-CLI
```

## Step 2: Install Dependencies

Using [Task](https://taskfile.dev/) (recommended):

```bash
task install
```

Or run the shell script directly:

```bash
chmod +x requirements.sh
./requirements.sh
```

This fetches all Go modules (Cobra, Viper, Supabase client, PTerm, Survey, Zap, Resty) and runs `go mod tidy`.

## Step 3: Configure Environment

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` and fill in the required values:

```env
EASYVPN_API_TOKEN=your-secret-token
EASYVPN_SUPABASE_URL=https://xxxxxxxxxxxx.supabase.co
EASYVPN_SUPABASE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

> **Security note:** Never commit `.env` to version control. It is listed in `.gitignore`.

If you run EasyVPN without a valid `.env`, the onboarding flow prompts you to create one or prints a template. See [Configuration](configuration.md) for all available variables.

## Step 4: Verify System Requirements

```bash
task run -- setup
# or: go run cmd/easyvpn/main.go setup
```

This checks that WireGuard tools and other platform dependencies are installed. On macOS, you can install `wireguard-tools` via Homebrew from the interactive menu (option 1).

## Step 5: Connect

### Interactive Mode

Run without arguments to open the main menu:

```bash
task run
```

On **Linux/Windows**, choose **Connect to VPN** (option 3). On **macOS**, choose **Export macOS WireGuard Config** (option 3) — see [Platform Support](platform-support.md#macos-darwin).

### Direct Command

```bash
task run -- connect
```

The connect flow:

1. Fetches active servers from Supabase
2. Presents a server selection prompt (`[region] name (ip)`)
3. Verifies node agent health
4. Generates a new WireGuard keypair
5. Provisions (or rotates) the peer on the server
6. Applies the tunnel configuration
7. Saves state to `~/.easyvpn/state.json`

## Step 6: Verify Connection

```bash
task run -- status
task run -- ip
```

`status` reconciles persisted state with the actual OS tunnel status. `ip` queries your public IP via the configured check URL (default: ipify.org).

## Step 7: Disconnect

```bash
task run -- disconnect
```

This disables the kill-switch (if active), tears down the tunnel, and clears connection state.

## macOS Users

macOS does not manage tunnels directly from the CLI. Instead, EasyVPN:

1. Provisions your peer on the server
2. Writes `~/Desktop/EasyVPN_macOS.conf`
3. Prints instructions to import the file into the official [WireGuard macOS app](https://www.wireguard.com/install/)

After importing, activate the tunnel from the WireGuard app.

## Next Steps

- [CLI Reference](cli-reference.md) — All commands and menu options
- [Configuration](configuration.md) — Optional settings and config directory
- [Troubleshooting](troubleshooting.md) — If something goes wrong
