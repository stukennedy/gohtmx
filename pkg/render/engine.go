// Package render provides a template engine for HTML fragment rendering.
package render

import (
	"bytes"
	"html/template"
	"io/fs"
	"sync"
)

// Engine manages template parsing and rendering.
type Engine struct {
	templates *template.Template
	funcs     template.FuncMap
	mu        sync.RWMutex
}

// New creates a new template engine with default functions.
func New() *Engine {
	e := &Engine{
		funcs: DefaultFuncs(),
	}
	return e
}

// AddFunc registers a custom template function.
// Must be called before loading templates.
func (e *Engine) AddFunc(name string, fn any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcs[name] = fn
}

// AddFuncs registers multiple template functions.
func (e *Engine) AddFuncs(funcs template.FuncMap) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for name, fn := range funcs {
		e.funcs[name] = fn
	}
}

// LoadFS loads templates from an embedded filesystem.
func (e *Engine) LoadFS(fsys fs.FS, patterns ...string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tmpl, err := template.New("").Funcs(e.funcs).ParseFS(fsys, patterns...)
	if err != nil {
		return err
	}
	e.templates = tmpl
	return nil
}

// LoadGlob loads templates matching a glob pattern.
func (e *Engine) LoadGlob(pattern string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tmpl, err := template.New("").Funcs(e.funcs).ParseGlob(pattern)
	if err != nil {
		return err
	}
	e.templates = tmpl
	return nil
}

// LoadFiles loads specific template files.
func (e *Engine) LoadFiles(filenames ...string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tmpl, err := template.New("").Funcs(e.funcs).ParseFiles(filenames...)
	if err != nil {
		return err
	}
	e.templates = tmpl
	return nil
}

// Parse parses a template string directly.
func (e *Engine) Parse(name, text string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.templates == nil {
		e.templates = template.New("").Funcs(e.funcs)
	}

	_, err := e.templates.New(name).Parse(text)
	return err
}

// Render executes a template and returns HTML string.
func (e *Engine) Render(name string, data any) (string, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.templates == nil {
		return "", &TemplateError{Name: name, Err: ErrNoTemplates}
	}

	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", &TemplateError{Name: name, Err: err}
	}
	return buf.String(), nil
}

// MustRender executes a template and panics on error.
func (e *Engine) MustRender(name string, data any) string {
	html, err := e.Render(name, data)
	if err != nil {
		panic(err)
	}
	return html
}

// Fragment renders a partial/component from the fragments directory.
// Prepends "fragments/" to the name.
func (e *Engine) Fragment(name string, data any) (string, error) {
	return e.Render("fragments/"+name, data)
}

// Component renders a reusable UI component.
// Prepends "components/" to the name.
func (e *Engine) Component(name string, data any) (string, error) {
	return e.Render("components/"+name, data)
}

// Page renders a full page template.
// Prepends "pages/" to the name.
func (e *Engine) Page(name string, data any) (string, error) {
	return e.Render("pages/"+name, data)
}

// Layout renders a layout template.
// Prepends "layouts/" to the name.
func (e *Engine) Layout(name string, data any) (string, error) {
	return e.Render("layouts/"+name, data)
}

// HasTemplate checks if a template exists.
func (e *Engine) HasTemplate(name string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.templates == nil {
		return false
	}
	return e.templates.Lookup(name) != nil
}

// Templates returns the list of defined template names.
func (e *Engine) Templates() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.templates == nil {
		return nil
	}

	var names []string
	for _, t := range e.templates.Templates() {
		if t.Name() != "" {
			names = append(names, t.Name())
		}
	}
	return names
}

// Clone creates a copy of the engine with the same templates and funcs.
func (e *Engine) Clone() (*Engine, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	clone := &Engine{
		funcs: make(template.FuncMap),
	}

	for k, v := range e.funcs {
		clone.funcs[k] = v
	}

	if e.templates != nil {
		cloned, err := e.templates.Clone()
		if err != nil {
			return nil, err
		}
		clone.templates = cloned
	}

	return clone, nil
}
