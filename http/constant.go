package http

type errorCode int

// defines server response error codes
const (
	ErrorCodeFailed  errorCode = -1
	ErrorCodeSuccess errorCode = iota
	ErrorCodeInternalError
	ErrorCodeThirdPartyError
	ErrorCodeMalformedMethod
	ErrorCodeValidationFailed
)

// HTTP headers
const (
	HeaderOrigin           = "Origin"
	HeaderAccept           = "Accept"
	HeaderContentType      = "Content-Type"
	HeaderAuthorization    = "Authorization"
	HeaderAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAllowCredentials = "Access-Control-Allow-Credentials"
)

// Content types
const (
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html"
	ContentTypeText = "text/plain"
	ContentTypeForm = "application/x-www-form-urlencoded"
)

type contextValidator string

// defines context key
const (
	ContextValidatorKey contextValidator = "validator"
)
