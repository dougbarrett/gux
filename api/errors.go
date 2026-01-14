package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Error represents an API error with HTTP status code
type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

// ErrorResponse is the JSON structure returned to clients
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError writes an API error as JSON response
func WriteError(w http.ResponseWriter, err error) {
	apiErr, ok := err.(*Error)
	if !ok {
		apiErr = &Error{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: err.Error(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorBody{
			Code:    apiErr.Code,
			Message: apiErr.Message,
		},
	})
}

// Common error constructors

func NotFound(message string) *Error {
	return &Error{Status: http.StatusNotFound, Code: "not_found", Message: message}
}

func NotFoundf(format string, args ...any) *Error {
	return NotFound(fmt.Sprintf(format, args...))
}

func BadRequest(message string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: "bad_request", Message: message}
}

func BadRequestf(format string, args ...any) *Error {
	return BadRequest(fmt.Sprintf(format, args...))
}

func Unauthorized(message string) *Error {
	return &Error{Status: http.StatusUnauthorized, Code: "unauthorized", Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Status: http.StatusForbidden, Code: "forbidden", Message: message}
}

func Conflict(message string) *Error {
	return &Error{Status: http.StatusConflict, Code: "conflict", Message: message}
}

func InternalError(message string) *Error {
	return &Error{Status: http.StatusInternalServerError, Code: "internal_error", Message: message}
}

func InternalErrorf(format string, args ...any) *Error {
	return InternalError(fmt.Sprintf(format, args...))
}
