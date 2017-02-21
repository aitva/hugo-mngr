package mngr

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strings"
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
		// Folder contains the name of the folder.
		Folder string
	}
)

var (
	validURLKey = validURLCtxKey(0)
)

// ValidURLFromCtx extract a ValidURL added by MakeValidURLMiddleware from a context.
func ValidURLFromCtx(ctx context.Context) (ValidURL, bool) {
	valid, ok := ctx.Value(validURLKey).(ValidURL)
	return valid, ok
}

// findFolder separate folder and file in the path.
// Folder will be empty if there is only a file.
func findFolder(path string) (file, folder string) {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		file = path
		return
	}
	file = path[i:]
	folder = path[:i]
	return
}

// MakeValidURLMiddleware create an URL validation middleware.
// When plugged, the returned middleware add a ValidURL to the request's context.
func MakeValidURLMiddleware() Middleware {
	validPath := regexp.MustCompile("^/([a-z]+)/([a-zA-Z0-9/]*[a-zA-Z0-9]+[a-zA-Z0-9.]*)$")
	return func(h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			m := validPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request: URL validation failed"))
				return http.StatusBadRequest, nil
			}
			file, folder := findFolder(m[2])
			ctx := r.Context()
			ctx = context.WithValue(ctx, validURLKey, ValidURL{
				Action: m[1],
				Value:  file,
				Folder: folder,
			})
			return h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MakeValidFolderMiddleware create an URL validation middleware.
// When plugged, the returned middleware will look for a valid URL and
// an existing folder on disk. It will also add a ValidURL to the request's
// context.
func MakeValidFolderMiddleware(dataPath string) Middleware {
	validPath := regexp.MustCompile("^/([a-z]+)/([a-zA-Z0-9/]*)$")
	return func(h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			m := validPath.FindStringSubmatch(r.URL.Path)
			path := m[2]
			if len(path) != 0 && path[len(path)-1] != '/' {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
				return http.StatusFound, nil
			}
			f, err := os.Stat(dataPath + "/" + path)
			if err != nil {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("not found"))
				return http.StatusNotFound, nil
			}
			if !f.IsDir() {
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request"))
				return http.StatusBadRequest, nil
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, validURLKey, ValidURL{
				Action: m[1],
				Folder: m[2],
			})
			return h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
