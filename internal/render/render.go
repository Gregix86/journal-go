package render

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	pages map[string]*template.Template
}

// New parses templates/base.html together with every page under templates/
// (and templates/admin/) so each page can be executed independently.
func New(dir string) (*Renderer, error) {
	base := filepath.Join(dir, "base.html")

	pageFiles, err := filepath.Glob(filepath.Join(dir, "*.html"))
	if err != nil {
		return nil, err
	}
	adminFiles, err := filepath.Glob(filepath.Join(dir, "admin", "*.html"))
	if err != nil {
		return nil, err
	}

	r := &Renderer{pages: make(map[string]*template.Template)}

	for _, f := range pageFiles {
		name := filepath.Base(f)
		if name == "base.html" {
			continue
		}
		tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(base, f)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f, err)
		}
		r.pages[name] = tmpl
	}
	for _, f := range adminFiles {
		name := "admin/" + filepath.Base(f)
		tmpl, err := template.New("base.html").Funcs(funcMap).ParseFiles(base, f)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", f, err)
		}
		r.pages[name] = tmpl
	}

	return r, nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
	tmpl, ok := r.pages[name]
	if !ok {
		http.Error(w, "template introuvable: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
