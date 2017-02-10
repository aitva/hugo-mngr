package hugomngr

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

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

type File struct {
	Name  string
	IsDir bool
}

// ViewInfo contains the minimum informations needed
// to render a Template.
type ViewInfo struct {
	Action string
	Page   *Page
	Files  []File
}

func renderTemplate(w http.ResponseWriter, tmpl string, v *ViewInfo) {
	err := templates.ExecuteTemplate(w, tmpl, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func MakeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
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
	v := &ViewInfo{
		Action: "index",
		Page: &Page{
			Filename: "/",
		},
		Files: append(dstFolders, dstFiles...),
	}
	renderTemplate(w, "index.html", v)
}

func ViewHandler(w http.ResponseWriter, r *http.Request, filename string) {
	p, err := loadPage(filename)
	if err != nil {
		http.Redirect(w, r, "/edit/"+filename, http.StatusFound)
		return
	}
	v := &ViewInfo{
		Action: "view",
		Page:   p,
	}
	renderTemplate(w, "view.html", v)
}

func EditHandler(w http.ResponseWriter, r *http.Request, filename string) {
	p, err := loadPage(filename)
	if err != nil {
		p = &Page{Filename: filename}
	}
	v := &ViewInfo{
		Action: "edit",
		Page:   p,
	}
	renderTemplate(w, "edit.html", v)
}

func SaveHandler(w http.ResponseWriter, r *http.Request, filename string) {
	body := r.FormValue("body")
	p := &Page{Filename: filename, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+filename, http.StatusFound)
}
