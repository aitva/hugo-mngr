package mngr

import (
	"io/ioutil"
	"os"
)

const pagesPath = "data"

// Page represet a wiki page.
type Page struct {
	TemplateInfo
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

func createFolder(name string) error {
	err := os.Mkdir(pagesPath+"/"+name, 0600)
	return err
}
