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

// StatusWriter is an http.ResponseWriter which
// captures the status set with WriteHeader.
type StatusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader is a redefinition of http.ResponseWriter.WriteHeader.
// This function allows us to capture the status set by an handler.
func (w *StatusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// fileHandler creates a fileserver and captures the response
// status for logging.
func fileHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	sw := &StatusWriter{ResponseWriter: w}
	h := http.FileServer(http.Dir("static"))
	h = http.StripPrefix("/static/", h)
	h.ServeHTTP(sw, r)
	return sw.status, nil
}

func main() {
	const addr = ":8080"

	log := mngr.MakeLogMiddleware(os.Stdout)
	tmpl := mngr.MakeTemplateMiddleware(tmplPath)
	valid := mngr.MakeValidURLMiddleware()
	index := log(tmpl(mngr.MakeIndexHandler(dataPath)))
	view := log(tmpl(valid(mngr.HandlerFunc(mngr.ViewHandler))))
	edit := log(tmpl(valid(mngr.HandlerFunc(mngr.EditHandler))))
	save := log(tmpl(valid(mngr.HandlerFunc(mngr.SaveHandler))))
	new := log(tmpl(mngr.HandlerFunc(mngr.NewHandler)))
	filesrv := log(mngr.HandlerFunc(fileHandler))

	http.Handle("/", index)
	http.Handle("/view/", view)
	http.Handle("/edit/", edit)
	http.Handle("/save/", save)
	http.Handle("/new/", new)
	http.Handle("/static/", filesrv)

	fmt.Println("Listening on " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
