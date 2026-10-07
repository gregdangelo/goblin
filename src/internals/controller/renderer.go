package controller

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
)

/*
We're going to want to embed all the HTML,CSS and JS within the binary file
a problem for another day
*/
type Renderer struct {
	templates map[string]*template.Template
	mu        sync.Mutex
	logger    *slog.Logger
}

func NewRenderer(logger *slog.Logger) (*Renderer, error) {
	r := Renderer{
		templates: make(map[string]*template.Template),
		logger:    logger,
	}

	funcs := template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				dict[values[i].(string)] = values[i+1]
			}
			return dict, nil
		},
	}
	common := template.Must(template.New("base").Funcs(funcs).ParseGlob("src/assets/html/layouts/*.html"))
	//components
	template.Must(common.ParseGlob("src/assets/html/partials/*.html"))

	templateDirs := []string{"src/assets/html/pages"}

	for _, folderpath := range templateDirs {
		var components []string
		filepath.Walk(folderpath, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && strings.Contains(path, "/partials/") {
				components = append(components, path)
			}
			return nil
		})
		// Walk through all directories matching the pattern
		err := filepath.Walk(folderpath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.Contains(path, "/partials/") {
				return nil
			}
			if filepath.Ext(path) != ".html" {
				return nil
			}

			// clone the common items
			tmpl, err := common.Clone()
			if err != nil {
				return err
			}
			if len(components) > 0 {
				tmpl = template.Must(tmpl.ParseFiles(components...))
			}
			tmpl = template.Must(tmpl.ParseFiles(path))

			rel := strings.TrimSuffix(path, filepath.Ext(path))

			rel = strings.TrimPrefix(rel, "src/assets/html/")

			rel = strings.TrimSpace(rel)
			logger.Debug("path", "rel", rel)
			r.templates[rel] = tmpl // Share same template set

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// debug
	logger.Debug("templates", "common", common, "templates", r.templates)

	return &r, nil
}

func (r *Renderer) Render(w io.Writer, templatePath string, templateName string, data interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tmpl, exists := r.templates[templatePath]
	if !exists {
		r.logger.Error("template %s not found", templatePath)
		return fmt.Errorf("template %s not found", templatePath)
	}

	return tmpl.ExecuteTemplate(w, templateName, data)
}
