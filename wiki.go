package main

import (
	"fmt"
	"io/ioutil"
	"os"
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

func main() {
	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample page.")}
	err := p1.save()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	p2, err := loadPage("TestPage")
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(p2.Body))
}
