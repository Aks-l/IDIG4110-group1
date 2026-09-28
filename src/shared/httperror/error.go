package httperror

const (
	ErrBadRequest          = "bad request"
	ErrUnauthorized        = "unauthorized"
	ErrForbidden           = "forbidden"
	ErrNotFound            = "resource not found"
	ErrMethodNotAllowed    = "method not allowed"
	ErrConflict            = "conflict"
	ErrGone                = "resource is gone"
	ErrUnprocessableEntity = "unprocessable entity"
	ErrTooManyRequests     = "too many requests"

	ErrInternalServerError = "internal server error"
	ErrBadGateway          = "bad gateway"
	ErrServiceUnavailable  = "service unavailable"
	ErrGatewayTimeout      = "gateway timeout"
)