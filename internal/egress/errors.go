package egress

import "fmt"

// Error codes for JS error handling.
const (
	CodeBlocked = "NET_BLOCKED" // Allowlist/IP rejected (not retryable)
	CodeTimeout = "NET_TIMEOUT" // Upstream timeout (not retryable)
	CodeLimit   = "NET_LIMIT"   // Concurrency limit hit (retryable)
	CodeBudget  = "NET_BUDGET"  // Insufficient time budget (retryable)
	CodeSize    = "NET_SIZE"    // Request/response body too large (not retryable)
	CodeError   = "NET_ERROR"   // Other network error (not retryable)
	CodeAuth    = "NET_AUTH"    // Secret not found or domain mismatch (not retryable)
	CodeRate    = "NET_RATE"    // Rate limited (retryable)
)

// EgressError is a structured error with a stable code for JS error handling.
type EgressError struct {
	Code      string
	Message   string
	Retryable bool
}

func (e *EgressError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// IsRetryableError returns true if the error is a retryable EgressError.
func IsRetryableError(err error) bool {
	if ee, ok := err.(*EgressError); ok {
		return ee.Retryable
	}
	return false
}

func errBlocked(msg string) *EgressError {
	return &EgressError{Code: CodeBlocked, Message: msg, Retryable: false}
}

func errTimeout(msg string) *EgressError {
	return &EgressError{Code: CodeTimeout, Message: msg, Retryable: false}
}

func errLimit(msg string) *EgressError {
	return &EgressError{Code: CodeLimit, Message: msg, Retryable: true}
}

func errBudget(msg string) *EgressError {
	return &EgressError{Code: CodeBudget, Message: msg, Retryable: true}
}

func errSize(msg string) *EgressError {
	return &EgressError{Code: CodeSize, Message: msg, Retryable: false}
}

func errNet(msg string) *EgressError {
	return &EgressError{Code: CodeError, Message: msg, Retryable: false}
}

func errAuth(msg string) *EgressError {
	return &EgressError{Code: CodeAuth, Message: msg, Retryable: false}
}

func errRate(msg string) *EgressError {
	return &EgressError{Code: CodeRate, Message: msg, Retryable: true}
}
