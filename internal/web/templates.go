package web

import (
	"embed"
	"fmt"
	"html/template"
	"log"
)

// TemplateManager manages HTML templates
type TemplateManager struct {
	tmpl *template.Template
	fs   embed.FS
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templatesFS embed.FS) *TemplateManager {
	return &TemplateManager{fs: templatesFS}
}

// LoadTemplates loads HTML templates from embedded filesystem
// This preserves the exact same logic from the original loadTemplates function
func (tm *TemplateManager) LoadTemplates() error {
	var err error
	tm.tmpl, err = template.ParseFS(tm.fs, "templates/*.html")
	if err != nil {
		return fmt.Errorf("error parsing templates: %w", err)
	}
	log.Println("HTML templates loaded.")
	return nil
}

// GetTemplate returns the loaded template
func (tm *TemplateManager) GetTemplate() *template.Template {
	return tm.tmpl
}
