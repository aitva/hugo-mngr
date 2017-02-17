package mngr

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

// MakeLogMiddleware create a logging middleware who wan be plugged into the
// default Go http.Server. The middleware traces every request and handle
// the response if mngr.Handler return 0 and an error.
func MakeLogMiddleware(out io.Writer) func(h Handler) http.Handler {
	return func(h Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := time.Now()
			code, err := h.ServeHTTP(w, r)
			if code == 0 && err != nil {
				code = http.StatusInternalServerError
				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(code)
				fmt.Fprintln(w, err)
			}
			elapsed := fmt.Sprintf("%0.3fs", time.Since(t).Seconds())
			fmt.Fprintln(out, r.RemoteAddr, elapsed, code, r.Method, r.URL.Path, err)
		})
	}
}

// MakeIndexHandler return an handler for the index page.
// The handler will list all the file present in dataPath.
func MakeIndexHandler(dataPath string) HandlerFunc {
	type File struct {
		Name  string
		IsDir bool
	}
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		files, err := ioutil.ReadDir(dataPath)
		if err != nil {
			return 0, err
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
		v := &struct {
			Page
			Files []File
		}{}
		v.Action = "index"
		v.Filename = "/"
		v.Files = append(dstFolders, dstFiles...)

		t, _ := TemplateFromCtx(r.Context())
		err = t.ExecuteTemplate(w, "index.html", v)
		return 200, err
	}
}

// ViewHandler is an handler use to display the content of a file.
func ViewHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	p, err := loadPage(valid.Value)
	if err != nil {
		http.Redirect(w, r, "/edit/"+valid.Value, http.StatusFound)
		return http.StatusFound, nil
	}
	p.Action = "view"
	t, _ := TemplateFromCtx(r.Context())
	err = t.ExecuteTemplate(w, "view.html", p)
	return 200, err
}

// EditHandler is an handler use to edit the content of a file.
func EditHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	p, err := loadPage(valid.Value)
	if err != nil {
		p = &Page{Filename: valid.Value}
	}
	p.Action = "edit"
	t, _ := TemplateFromCtx(r.Context())
	err = t.ExecuteTemplate(w, "edit.html", p)
	return 200, err
}

// SaveHandler is an handler use to save the content of a page in a file.
func SaveHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	body := r.FormValue("body")
	p := &Page{Filename: valid.Value, Body: []byte(body)}
	err := p.save()
	if err != nil {
		return 0, err
	}
	http.Redirect(w, r, "/view/"+valid.Value, http.StatusFound)
	return http.StatusFound, nil
}

// FolderHandler is a HandlerFunc use to create new folder.
func FolderHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	err := createFolder(valid.Value)
	if err != nil {
		return 0, err
	}
	// TODO: redirect to folder listing.
	http.Redirect(w, r, "/index/", http.StatusFound)
	return http.StatusFound, nil
}

// MakeCreateHandler return an HandlerFunc which deals with file and folder creation.
func MakeCreateHandler() HandlerFunc {
	validFilename := regexp.MustCompile("^[a-zA-Z0-9]+[a-zA-Z0-9.]*$")
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		valid, _ := ValidURLFromCtx(r.Context())
		if valid.Value != "file" && valid.Value != "folder" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
			return http.StatusBadRequest, nil
		}

		name := r.URL.Query().Get("name")
		if name == "" {
			p := &Page{
				TemplateInfo: TemplateInfo{
					Action: "new " + valid.Value,
					Value:  valid.Value,
				},
				Filename: "",
			}
			t, _ := TemplateFromCtx(r.Context())
			err := t.ExecuteTemplate(w, "new.html", p)
			return 200, err
		}

		if !validFilename.MatchString(name) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request: invalid name"))
			return http.StatusBadRequest, nil
		}

		p := "/edit/" + name
		if valid.Value == "folder" {
			p = "/folder/" + name
		}
		http.Redirect(w, r, p, http.StatusFound)
		return http.StatusFound, nil
	}
}
