package hugomngr

import (
	"io/ioutil"
	"net/http"
)

type (
	// File represent a File on disk. It is use when rendering templates.
	// TODO: remove
	File struct {
		Name  string
		IsDir bool
	}

	// ViewInfo contains the minimum informations needed
	// to render a Template.
	ViewInfo struct {
		Action string
		Page   *Page
		Files  []File
	}
)

// MakeIndexHandler return an handler for the index page.
// The handler will list all the file present in dataPath.
func MakeIndexHandler(dataPath string) HandlerFunc {
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
		v := &ViewInfo{
			Action: "index",
			Page: &Page{
				Filename: "/",
			},
			Files: append(dstFolders, dstFiles...),
		}
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
	v := &ViewInfo{
		Action: "view",
		Page:   p,
	}
	t, _ := TemplateFromCtx(r.Context())
	err = t.ExecuteTemplate(w, "view.html", v)
	return 200, err
}

// EditHandler is an handler use to edit the content of a file.
func EditHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	valid, _ := ValidURLFromCtx(r.Context())
	p, err := loadPage(valid.Value)
	if err != nil {
		p = &Page{Filename: valid.Value}
	}
	v := &ViewInfo{
		Action: "edit",
		Page:   p,
	}
	t, _ := TemplateFromCtx(r.Context())
	err = t.ExecuteTemplate(w, "edit.html", v)
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
