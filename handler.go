package mngr

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// MakeLogMiddleware create a logging middleware who wan be plugged into the
// default Go http.Server. The middleware traces every request and handle
// the response if mngr.Handler return 0 and an error.
func MakeLogMiddleware(out io.Writer) func(h Handler) http.HandlerFunc {
	return func(h Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
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
		}
	}
}

func filterFiles(fInfos []os.FileInfo) (files, folders []string) {
	files = make([]string, 0, len(fInfos))
	folders = make([]string, 0, len(fInfos))
	for _, f := range fInfos {
		name := f.Name()
		// Skip files starting with a dot.
		if name[0] == '.' {
			continue
		}
		if f.IsDir() {
			folders = append(folders, name)
		} else {
			files = append(files, name)
		}
	}
	return
}

// MakeListHandler return an handler wich list folder's content.
// The handler will list all the file present in dataPath.
func MakeListHandler(dataPath string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		valid, _ := ValidURLFromCtx(r.Context())
		fInfos, err := ioutil.ReadDir(dataPath + "/" + valid.Folder)
		if err != nil {
			return 0, err
		}
		files, folders := filterFiles(fInfos)
		v := &struct {
			Page
			Files   []string
			Folders []string
		}{
			Files:   files,
			Folders: folders,
		}
		v.TemplateInfo = TemplateInfo(valid)
		v.Filename = valid.Folder

		t, _ := TemplateFromCtx(r.Context())
		err = t.ExecuteTemplate(w, "list.html", v)
		return 200, err
	}
}

// ViewHandler is an handler use to display the content of a file.
func ViewHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	path := valid.Folder + valid.Value
	p, err := loadPage(path)
	if err != nil {
		http.Redirect(w, r, "/edit/"+path, http.StatusFound)
		return http.StatusFound, nil
	}
	p.TemplateInfo = TemplateInfo(valid)
	t, _ := TemplateFromCtx(r.Context())
	err = t.ExecuteTemplate(w, "view.html", p)
	return 200, err
}

// EditHandler is an handler use to edit the content of a file.
func EditHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	path := valid.Folder + valid.Value
	p, err := loadPage(path)
	if err != nil {
		p = &Page{Filename: path}
	}
	p.TemplateInfo = TemplateInfo(valid)
	t, _ := TemplateFromCtx(r.Context())
	err = t.ExecuteTemplate(w, "edit.html", p)
	return 200, err
}

// SaveHandler is an handler use to save the content of a page in a file.
func SaveHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	body := r.FormValue("body")
	path := valid.Folder + valid.Value
	p := &Page{Filename: path, Body: []byte(body)}
	p.TemplateInfo = TemplateInfo(valid)
	err := p.save()
	if err != nil {
		return 0, err
	}
	http.Redirect(w, r, "/view/"+path, http.StatusFound)
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
	http.Redirect(w, r, "/", http.StatusFound)
	return http.StatusFound, nil
}

// MakeNewHandler return an HandlerFunc which deals with file and folder creation.
func MakeNewHandler() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		valid, _ := ValidURLFromCtx(r.Context())
		if valid.Value != "file" && valid.Value != "folder" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("bad request"))
			return http.StatusBadRequest, nil
		}

		name := r.URL.Query().Get("name")
		path := r.URL.Query().Get("path")
		if name != "" {
			path = path + name
			url := "/edit/" + path
			if valid.Value == "folder" {
				url = "/folder/" + path
			}
			http.Redirect(w, r, url, http.StatusFound)
			return http.StatusFound, nil
		}

		p := &struct {
			Page
			Path string
		}{
			Path: path,
		}
		p.TemplateInfo = TemplateInfo(valid)
		p.Action = "new " + valid.Value
		p.Filename = path
		p.Folder = path
		t, _ := TemplateFromCtx(r.Context())
		err := t.ExecuteTemplate(w, "new.html", p)
		return 200, err
	}
}
