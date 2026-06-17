package models

import "fmt"

// ErrorCode is a unique machine-readable identifier for specific errors
type ErrorCode string

const (
	// System Errors
	ErrPermissionDenied  ErrorCode = "ERR_PERMISSION_DENIED"
	ErrMissingDependency ErrorCode = "ERR_MISSING_DEPENDENCY"
	ErrInternal          ErrorCode = "ERR_INTERNAL"

	// Configuration Errors
	ErrConfigMissing ErrorCode = "ERR_CONFIG_MISSING"
	ErrInvalidInput  ErrorCode = "ERR_INVALID_INPUT"

	// Network/API Errors
	ErrAuthFailed       ErrorCode = "ERR_AUTH_FAILED"
	ErrDiscoveryFailed  ErrorCode = "ERR_DISCOVERY_FAILED"
	ErrAgentUnreachable ErrorCode = "ERR_AGENT_UNREACHABLE"

	// VPN Errors
	ErrTunnelFailed   ErrorCode = "ERR_TUNNEL_FAILED"
	ErrKillSwitchFail ErrorCode = "ERR_KILLSWITCH_FAILED"
)

// EasyVPNError is our custom "beast-mode" error structure
type EasyVPNError struct {
	Code        ErrorCode // Machine-readable code
	Message     string    // Human-readable explanation
	Remediation string    // Actionable steps for the user to fix the issue
	Internal    error     // The original underlying error (for logging)
}

// Error implements the standard Go error interface
func (e *EasyVPNError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s (Internal: %v)", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewError is a helper to create a fresh custom error
func NewError(code ErrorCode, msg string, remediation string, internal error) *EasyVPNError {
	return &EasyVPNError{
		Code:        code,
		Message:     msg,
		Remediation: remediation,
		Internal:    internal,
	}
}

// WrapError is used to wrap an existing system error with our custom context
func WrapError(code ErrorCode, msg string, remediation string, internal error) *EasyVPNError {
	return &EasyVPNError{
		Code:        code,
		Message:     msg,
		Remediation: remediation,
		Internal:    internal,
	}
}
