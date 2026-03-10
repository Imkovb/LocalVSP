package view

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"
)

var (
	templates     *template.Template
	templatesErr  error
	templatesOnce sync.Once
)

func loadTemplates() (*template.Template, error) {
	templatesOnce.Do(func() {
		funcMap := template.FuncMap{
			"helpHTML": func(s string) template.HTML { return template.HTML(s) },
		}
		templates, templatesErr = template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html")
	})

	return templates, templatesErr
}

func Render(w http.ResponseWriter, name string, data any) {
	tpl, err := loadTemplates()
	if err != nil {
		http.Error(w, fmt.Sprintf("templates not loaded: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}