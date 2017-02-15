package mngr

import (
	"context"
	"net/http"
	"regexp"
)

type (
	validURLCtxKey int

	// ValidURL represent a valid URL for our application.
	ValidURL struct {
		// Action represent the action requested in the URL.
		// This field only support the values: edit, view, save
		Action string
		// Value contains the object on wich this action apply.
		// It must contain something of the form: 'filename.ext' of 'filename'
		Value string
	}
)

var validURLKey = validURLCtxKey(0)

// ValidURLFromCtx extract a ValidURL added by MakeValidURLMiddleware from a context.
func ValidURLFromCtx(ctx context.Context) (ValidURL, bool) {
	valid, ok := ctx.Value(validURLKey).(ValidURL)
	return valid, ok
}

// MakeValidURLMiddleware create an URL validation middleware.
// When plugged, the returned middleware add a ValidURL to the request's context.
func MakeValidURLMiddleware() Middleware {
	validPath := regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+[a-zA-Z0-9.]*)$")
	return func(h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request: URL validation failed"))
				return http.StatusBadRequest, nil
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, validURLKey, ValidURL{
				Action: m[1],
				Value:  m[2],
			})
			return h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
