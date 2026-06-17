#!/bin/bash

echo "🚀 Initializing EasyVPN Project Structure..."

# 1. Create Directory Structure
mkdir -p cmd/easyvpn \
internal/api \
internal/config \
internal/core \
internal/platform \
internal/state \
internal/ui \
pkg/models \
pkg/utils \
scripts \
build

# 2. Create Placeholder Files
touch cmd/easyvpn/main.go
touch internal/api/supabase.go
touch internal/api/agent.go
touch internal/config/config.go
touch internal/core/reconciler.go
touch internal/core/vpn.go
touch internal/platform/adapter.go
touch internal/platform/linux.go
touch internal/platform/windows.go
touch internal/platform/darwin.go
touch internal/state/store.go
touch internal/ui/menu.go
touch internal/ui/spinner.go
touch pkg/models/server.go
touch pkg/models/errors.go
touch pkg/utils/logger.go
touch pkg/utils/validator.go
touch Taskfile.yaml
touch .goreleaser.yaml

# 3. Populate .gitignore
cat <<EOT > .gitignore
# Binaries
bin/
dist/
easyvpn
easyvpn.exe

# Local Config & Secrets
.env
.env.local
config.json
state.json

# Dependencies
vendor/

# IDEs
.vscode/
.idea/
*.swp
EOT

# 4. Populate .env.example
cat <<EOT > .env.example
# --- REQUIRED ---
EASYVPN_API_TOKEN=
EASYVPN_SUPABASE_URL=
EASYVPN_SUPABASE_ANON_KEY=

# --- OPTIONAL ---
EASYVPN_NODE_API_BASE_URL=
EASYVPN_DEFAULT_REGION=
EASYVPN_PUBLIC_IP_CHECK_URL=https://api.ipify.org?format=json
EASYVPN_LOG_LEVEL=info
EASYVPN_ALLOWED_IPS_DEFAULT=0.0.0.0/0,::/0
EASYVPN_DNS_DEFAULT=1.1.1.1
EOT

echo "✅ Structure created successfully."