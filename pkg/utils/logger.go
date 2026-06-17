package utils

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger
var sensitiveStrings []string

// Initialize sets up the global logger with secret redaction
func Initialize(level string, secrets []string) {
	sensitiveStrings = secrets

	// 1. Define Log Level
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zap.DebugLevel
	case "warn":
		zapLevel = zap.WarnLevel
	case "error":
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	// 2. Build Encoder Config (The "Face" of the logs)
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Adds colors to terminal
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 3. Create the Core with Redaction
	// We wrap the standard console output with our custom redactor
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.Lock(os.Stdout),
		zapLevel,
	)

	// Wrap the core with our secret-scrubbing logic
	redactedCore := &redactingCore{
		Core: core,
	}

	// 4. Initialize the Global Logger
	logger := zap.New(redactedCore, zap.AddCaller())
	Log = logger.Sugar()
}

// redactingCore is a custom wrapper that scrubs sensitive data before it hits the console
type redactingCore struct {
	zapcore.Core
}

func (c *redactingCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// Scrub message
	ent.Message = scrub(ent.Message)

	// Scrub fields (metadata)
	for i := range fields {
		if fields[i].Type == zapcore.StringType {
			fields[i].String = scrub(fields[i].String)
		}
	}
	return c.Core.Write(ent, fields)
}

func scrub(text string) string {
	for _, secret := range sensitiveStrings {
		if secret != "" && strings.Contains(text, secret) {
			text = strings.ReplaceAll(text, secret, "[REDACTED]")
		}
	}
	return text
}

// Sync flushes any buffered log entries (Call this on app exit)
func Sync() {
	_ = Log.Sync()
}
