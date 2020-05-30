package domain

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Lists of Application Error Codes
var (
	AuthenticationFailCode   = 0x0041
	UnknownResourceCode      = 0x0044
	InvalidParamCode         = 0x0041
	OperationUnsupportedCode = 0x0043
	InternalErrorCode        = 0x0050
)

// AppError implements error containing application
// related error for debugging/logging purposes
type AppError struct {
	httpCode int
	code     int
	cause    error
	Msg      string `json:"errorMsg"`
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return e.Msg + ": " + e.cause.Error()
	}
	return e.Msg
}

// OK creates error structure that represent OK
func OK() error { return &AppError{httpCode: 200, Msg: "Success"} }

// Cause returns the underlying error cause of the wrapped error
func (e *AppError) Cause() error { return e.cause }

// Message returns the summarised error message for client
func (e *AppError) Message() string { return e.Msg }

// Code returns application error identifier (the error code)
func (e *AppError) Code() int { return e.code }

// HTTPCode returns associated http status code for the error
func (e *AppError) HTTPCode() int {
	if e.httpCode == 0 {
		return 200
	}
	return e.httpCode
}

// Wrap returns an app error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func (e *AppError) Wrap(err error, debug string) error {
	if err == nil {
		return nil
	}

	cause := errors.Wrap(err, debug)
	return &AppError{
		cause:    cause,
		code:     e.code,
		httpCode: e.httpCode,
		Msg:      e.Msg,
	}
}

// Wrapf returns an app error annotating err with stack trace
// at the point Wrapf is called, and formats the supplied message
func (e *AppError) Wrapf(err error, debugf string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	cause := errors.Wrap(err, fmt.Sprintf(debugf, args...))
	return &AppError{
		cause:    cause,
		code:     e.code,
		httpCode: e.httpCode,
		Msg:      e.Msg,
	}
}

// WithMessage append message to client with additional information.
func (e *AppError) WithMessage(message string) error {
	return &AppError{
		httpCode: e.httpCode,
		code:     e.code,
		cause:    e.cause,
		Msg:      e.Msg + ": " + message,
	}
}

// WithMessagef append message to client with additonal
// the format-specified information.
func (e *AppError) WithMessagef(messagef string, args ...interface{}) error {
	return &AppError{
		httpCode: e.httpCode,
		code:     e.code,
		cause:    e.cause,
		Msg:      e.Msg + ": " + fmt.Sprintf(messagef, args...),
	}
}

// ErrInternalServer translates exactly to shit happened
var ErrInternalServer = AppError{
	httpCode: http.StatusInternalServerError,
	code:     InternalErrorCode,
	Msg:      "Internal Server Error",
}

// ErrAuthenticationFail unrecognised user or token keys
var ErrAuthenticationFail = AppError{
	httpCode: http.StatusUnauthorized,
	code:     AuthenticationFailCode,
	Msg:      "Invalid credentials or unrecognized keys",
}

// ErrOperationNotSupported returned when user have no permission
// or hasn't pay the bill
var ErrOperationNotSupported = AppError{
	httpCode: http.StatusForbidden,
	code:     OperationUnsupportedCode,
	Msg:      "Insufficient Permission Required",
}

// ErrBadParameters returned when user have no permission
// or hasn't pay the bill
var ErrBadParameters = AppError{
	httpCode: http.StatusBadRequest,
	code:     InvalidParamCode,
	Msg:      "Insufficient Permission Required",
}

// ErrUnknownResource due to resource does not exist or
// not publicly available
var ErrUnknownResource = AppError{
	httpCode: http.StatusNotFound,
	code:     UnknownResourceCode,
	Msg:      "Requested resource not available",
}
