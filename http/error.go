package http

// BadRequestError define http bad request error
type BadRequestError struct {
	error
}

// ValidationError define http validation error
type ValidationError struct {
	error
}
