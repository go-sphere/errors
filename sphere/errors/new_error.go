package errors

import (
	stderrors "errors"
	"net/http"
)

// NewError creates an error carrying an HTTP status, an application code and a
// user-facing message, wrapping err. If err is nil, a default error derived from
// status is used: its text is http.StatusText(status), or "Unknown error" when
// status has no standard text.
//
// It is the default constructor called by the Join and JoinWithMessage helpers
// that protoc-gen-sphere-errors generates, and has the signature the plugin's
// new_errors_func flag requires. The returned value exposes its fields through
// the methods GetStatus() int32, GetCode() int32 and GetMessage() string, which
// is how github.com/go-sphere/httpx (and any other adapter) classifies errors,
// so code defining error enums does not need to import an HTTP adapter layer.
//
// Semantics match httpx.NewError: Error() returns err.Error(), Unwrap() returns
// err, and message is stored as given (no fallback is applied).
func NewError(status, code int32, message string, err error) error {
	if err == nil {
		err = statusError(status)
	}
	return &statusCodeError{
		err:     err,
		status:  status,
		code:    code,
		message: message,
	}
}

// statusCodeError is the concrete type returned by NewError. It stays
// unexported so callers depend on the GetStatus/GetCode/GetMessage methods
// rather than on the concrete type.
type statusCodeError struct {
	err     error
	status  int32
	code    int32
	message string
}

// Error returns the text of the wrapped error.
func (e *statusCodeError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped error, so errors.Is and errors.As see through it.
func (e *statusCodeError) Unwrap() error {
	return e.err
}

// GetStatus returns the HTTP status code.
func (e *statusCodeError) GetStatus() int32 {
	return e.status
}

// GetCode returns the application error code.
func (e *statusCodeError) GetCode() int32 {
	return e.code
}

// GetMessage returns the user-facing message.
func (e *statusCodeError) GetMessage() string {
	return e.message
}

func statusError(status int32) error {
	msg := http.StatusText(int(status))
	if msg == "" {
		msg = "Unknown error"
	}
	return stderrors.New(msg)
}
