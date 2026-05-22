package constants

type ctxKey string

const (
	ID = "id"

	MsgInvalidID        = "invalid ID"
	MsgInvalidJSONBody  = "invalid JSON body"
	MsgInternalError    = "internal server error"
	MsgRequestBodyLarge = "request body to large"
	MsgBadRequest       = "bad request"

	RequestIDKey    ctxKey = "request_id"
	HeaderRequestID        = "X-Request-Id"
)
