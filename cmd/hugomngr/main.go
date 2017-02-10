package main

import (
	"net/http"

	"fmt"
	"os"

	"github.com/aitva/hugomngr"
)

func main() {
	const addr = ":8080"
	http.HandleFunc("/", hugomngr.IndexHandler)
	http.HandleFunc("/view/", hugomngr.MakeHandler(hugomngr.ViewHandler))
	http.HandleFunc("/edit/", hugomngr.MakeHandler(hugomngr.EditHandler))
	http.HandleFunc("/save/", hugomngr.MakeHandler(hugomngr.SaveHandler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Listening on " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
