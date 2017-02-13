package hugomngr

import (
	"context"
	"html/template"
	"net/http"
	"strings"

	"github.com/russross/blackfriday"
)

type templateCtxKey int

var templateKey = templateCtxKey(0)

// TemplateFromCtx extract templates added by MakeTemplateMiddleware to a context.
func TemplateFromCtx(c context.Context) (*template.Template, bool) {
	t, ok := c.Value(templateKey).(*template.Template)
	return t, ok
}

// MakeTemplateMiddleware load an compile all templates located in 'path/*.html' and 'path/partial/*.html'.
// When plugged, the returned middleware add templates to the request's context.
func MakeTemplateMiddleware(path string) Middleware {
	var tmplFunc = template.FuncMap{
		"renderMD": func(data []byte) template.HTML {
			return template.HTML(blackfriday.MarkdownCommon(data))
		},
		"fmtTitle": func(title string) string {
			return strings.Title(title)
		},
	}
	templates := template.Must(template.New("main").Funcs(tmplFunc).ParseGlob(path + "/*.html"))
	templates = template.Must(templates.ParseGlob(path + "/partial/*.html"))

	return func(h Handler) Handler {
		return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, templateKey, templates)
			return h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
