#!/bin/bash

echo "📦 Installing Production Dependencies..."

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "❌ Error: Go is not installed. Please install Go first."
    exit 1
fi

# 1. CLI Framework & Configuration
echo "📥 Fetching Cobra (CLI) and Viper (Config)..."
go get github.com/spf13/cobra
go get github.com/spf13/viper

# 2. Interactive UI Components
echo "📥 Fetching Survey (Prompts) and PTerm (Styling)..."
go get github.com/AlecAivazis/survey/v2
go get github.com/pterm/pterm

# 3. Logging & Error Handling
echo "📥 Fetching Zap (Structured Logging)..."
go get go.uber.org/zap

# 4. Networking & Supabase
echo "📥 Fetching Supabase-Go & HTTP Utilities..."
go get github.com/supabase-community/supabase-go
go get github.com/go-resty/resty/v2

# 5. OS Execution (Safe shell commands)
echo "📥 Fetching Execa-style command runner..."
go get github.com/codeskyblue/go-sh

# Finalize
echo "🧹 Tying up go.mod..."
go mod tidy

echo "✅ All dependencies installed and go.mod is clean."