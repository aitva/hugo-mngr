package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"fmt"
	"os"

	"github.com/russross/blackfriday"
)

const (
	pagesPath = "data"
	tmplPath  = "tmpl"
)

var tmplFunc = template.FuncMap{
	"renderMD": func(data []byte) template.HTML {
		return template.HTML(blackfriday.MarkdownCommon(data))
	},
	"fmtTitle": func(title string) string {
		return strings.Title(title)
	},
}

var (
	templates = template.Must(template.New("main").Funcs(tmplFunc).ParseGlob(tmplPath + "/*.html"))
	validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
)

// Page represet a wiki page.
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := pagesPath + "/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := pagesPath + "/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{
		Title: title,
		Body:  body,
	}, nil
}

type View struct {
	Action string
	Page   *Page
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
		fmt.Println(r.RemoteAddr, r.Method, r.URL.Path)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, v *View) {
	err := templates.ExecuteTemplate(w, tmpl, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/welcome", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	v := &View{
		Action: "view",
		Page:   p,
	}
	renderTemplate(w, "view.html", v)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	v := &View{
		Action: "edit",
		Page:   p,
	}
	renderTemplate(w, "edit.html", v)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	const addr = ":8080"
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Listening on " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
