package render

import (
	"bytes"
	"context"
	"io"

	"github.com/a-h/templ"
)

// TemplRenderer wraps templ components for use with the router.
type TemplRenderer struct {
	ctx context.Context
}

// NewTemplRenderer creates a new templ renderer.
func NewTemplRenderer() *TemplRenderer {
	return &TemplRenderer{
		ctx: context.Background(),
	}
}

// WithContext returns a renderer with the given context.
func (r *TemplRenderer) WithContext(ctx context.Context) *TemplRenderer {
	return &TemplRenderer{ctx: ctx}
}

// Render renders a templ component to a string.
func (r *TemplRenderer) Render(component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(r.ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// MustRender renders a templ component, panics on error.
func (r *TemplRenderer) MustRender(component templ.Component) string {
	html, err := r.Render(component)
	if err != nil {
		panic(err)
	}
	return html
}

// RenderTo renders a templ component to a writer.
func (r *TemplRenderer) RenderTo(w io.Writer, component templ.Component) error {
	return component.Render(r.ctx, w)
}

// RenderComponent is a convenience function to render a templ component.
func RenderComponent(component templ.Component) (string, error) {
	return NewTemplRenderer().Render(component)
}

// MustRenderComponent renders a templ component, panics on error.
func MustRenderComponent(component templ.Component) string {
	return NewTemplRenderer().MustRender(component)
}

// TemplFunc wraps a templ component function for use as a FragmentHandler.
// Example:
//
//	router.GET("/", render.TemplFunc(func(ctx *router.Context) templ.Component {
//	    return pages.Home(data)
//	}))
type TemplFunc func(ctx context.Context) templ.Component

// TemplHandler returns a fragment handler that renders a templ component.
func TemplHandler(component templ.Component) func() (string, error) {
	return func() (string, error) {
		return RenderComponent(component)
	}
}
