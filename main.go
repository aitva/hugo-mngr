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
	templates = template.Must(
		template.Must(
			template.New("main").
				Funcs(tmplFunc).
				ParseGlob(tmplPath + "/*.html")).
			ParseGlob(tmplPath + "/partial/*.html"))
	validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+[a-zA-Z0-9.]*)$")
)

// Page represet a wiki page.
type Page struct {
	Filename string
	Body     []byte
}

func (p *Page) save() error {
	path := pagesPath + "/" + p.Filename
	return ioutil.WriteFile(path, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	path := pagesPath + "/" + title
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &Page{
		Filename: title,
		Body:     body,
	}, nil
}

type File struct {
	Name  string
	IsDir bool
}

type View struct {
	Action string
	Page   *Page
	Files  []File
}

func renderTemplate(w http.ResponseWriter, tmpl string, v *View) {
	err := templates.ExecuteTemplate(w, tmpl, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(pagesPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dstFiles := make([]File, 0, len(files))
	dstFolders := make([]File, 0, len(files))
	for _, f := range files {
		name := f.Name()
		if name[0] == '.' {
			continue
		}
		f := File{
			Name:  name,
			IsDir: f.IsDir(),
		}
		if f.IsDir {
			dstFolders = append(dstFolders, f)
		} else {
			dstFiles = append(dstFiles, f)
		}
	}
	v := &View{
		Action: "index",
		Page: &Page{
			Filename: "/",
		},
		Files: append(dstFolders, dstFiles...),
	}
	renderTemplate(w, "index.html", v)
}

func viewHandler(w http.ResponseWriter, r *http.Request, filename string) {
	p, err := loadPage(filename)
	if err != nil {
		http.Redirect(w, r, "/edit/"+filename, http.StatusFound)
		return
	}
	v := &View{
		Action: "view",
		Page:   p,
	}
	renderTemplate(w, "view.html", v)
}

func editHandler(w http.ResponseWriter, r *http.Request, filename string) {
	p, err := loadPage(filename)
	if err != nil {
		p = &Page{Filename: filename}
	}
	v := &View{
		Action: "edit",
		Page:   p,
	}
	renderTemplate(w, "edit.html", v)
}

func saveHandler(w http.ResponseWriter, r *http.Request, filename string) {
	body := r.FormValue("body")
	p := &Page{Filename: filename, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+filename, http.StatusFound)
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
