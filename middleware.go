package mngr

import "net/http"

// The code in this file is extracted from
// github.com/mholt/caddy/caddyhttp/httpserver/middleware.go
type (
	// Middleware is the middle layer which represents the traditional
	// idea of middleware: it chains one Handler to the next by being
	// passed the next Handler in the chain.
	Middleware func(Handler) Handler

	// Handler is like http.Handler except ServeHTTP may return a status
	// code and/or error.
	//
	// If ServeHTTP writes the response header, it should return a status
	// code of 0. This signals to other handlers before it that the response
	// is already handled, and that they should not write to it also. Keep
	// in mind that writing to the response body writes the header, too.
	//
	// If ServeHTTP encounters an error, it should return the error value
	// so it can be logged by designated error-handling middleware.
	//
	// If writing a response after calling the next ServeHTTP method, the
	// returned status code SHOULD be used when writing the response.
	//
	// If handling errors after calling the next ServeHTTP method, the
	// returned error value SHOULD be logged or handled accordingly.
	//
	// Otherwise, return values should be propagated down the middleware
	// chain by returning them unchanged.
	Handler interface {
		ServeHTTP(http.ResponseWriter, *http.Request) (int, error)
	}

	// HandlerFunc is a convenience type like http.HandlerFunc, except
	// ServeHTTP returns a status code and an error. See Handler
	// documentation for more information.
	HandlerFunc func(http.ResponseWriter, *http.Request) (int, error)
)

// ServeHTTP implements the Handler interface.
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	return f(w, r)
}
