package mngr

import (
	"io/ioutil"
	"os"
)

const pagesPath = "data"

// Page represet a wiki page.
type Page struct {
	TemplateInfo
	Path     string
	Filename string
	Body     []byte
}

func (p *Page) save() error {
	path := pagesPath + "/" + p.Path
	return ioutil.WriteFile(path, p.Body, 0600)
}

func PagePathFromValidURL(v ValidURL) string {
	return v.Dir + v.Value
}

func LoadPage(v ValidURL) (*Page, error) {
	path := PagePathFromValidURL(v)
	body, err := ioutil.ReadFile(pagesPath + "/" + path)
	if err != nil {
		return nil, err
	}
	return &Page{
		TemplateInfo: NewTemplateFromValidURL(v),
		Path:         path,
		Filename:     v.Value,
		Body:         body,
	}, nil
}

func NewPage(v ValidURL, body []byte) *Page {
	return &Page{
		TemplateInfo: NewTemplateFromValidURL(v),
		Path:         PagePathFromValidURL(v),
		Filename:     v.Value,
		Body:         body,
	}
}

func NewFolder(v ValidURL) error {
	path := pagesPath + "/" + v.Dir + v.Value
	err := os.Mkdir(path, 0700)
	return err
}
