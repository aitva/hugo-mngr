package mngr

import "io/ioutil"

const pagesPath = "data"

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
