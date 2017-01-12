package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

// Page represet a wiki page.
type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{
		Title: title,
		Body:  body,
	}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, _ := template.ParseFiles(tmpl)
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, p)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(title)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(400)
		fmt.Fprintf(w, "fail to load page: %v", title)
		return
	}
	renderTemplate(w, "view.html", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit.html", p)
}

func main() {
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	//http.HandleFunc("/save/", saveHandler)
	http.ListenAndServe(":8080", nil)
}
