package main

import (
	"net/http"

	"fmt"
	"os"

	"github.com/aitva/mngr"
)

const (
	dataPath = "data"
	tmplPath = "tmpl"
)

type StatusWriter struct {
	http.ResponseWriter
	status int
}

func (w *StatusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func log(h hugomngr.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code, err := h.ServeHTTP(w, r)
		if code == 0 && err != nil {
			code = http.StatusInternalServerError
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(code)
			fmt.Fprintln(w, err)
		}
		out := os.Stdout
		if err != nil {
			out = os.Stderr
		}
		fmt.Fprintln(out, r.RemoteAddr, code, r.Method, r.URL.Path, err)
	}
}

func main() {
	const addr = ":8080"
	tmpl := hugomngr.MakeTemplateMiddleware(tmplPath)
	valid := hugomngr.MakeValidURLMiddleware()
	index := log(tmpl(hugomngr.MakeIndexHandler(dataPath)))
	view := log(tmpl(valid(hugomngr.HandlerFunc(hugomngr.ViewHandler))))
	edit := log(tmpl(valid(hugomngr.HandlerFunc(hugomngr.EditHandler))))
	save := log(tmpl(valid(hugomngr.HandlerFunc(hugomngr.SaveHandler))))
	filesrv := log(hugomngr.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
		sw := &StatusWriter{ResponseWriter: w}
		h := http.FileServer(http.Dir("static"))
		h = http.StripPrefix("/static/", h)
		h.ServeHTTP(sw, r)
		return sw.status, nil
	}))
	http.HandleFunc("/", index)
	http.HandleFunc("/view/", view)
	http.HandleFunc("/edit/", edit)
	http.HandleFunc("/save/", save)
	http.Handle("/static/", filesrv)
	fmt.Println("Listening on " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
