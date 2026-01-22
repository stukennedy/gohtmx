package render

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
)

var (
	// ErrNoTemplates is returned when no templates have been loaded.
	ErrNoTemplates = errors.New("no templates loaded")
)

// TemplateError wraps template execution errors.
type TemplateError struct {
	Name string
	Err  error
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf("template %q: %v", e.Name, e.Err)
}

func (e *TemplateError) Unwrap() error {
	return e.Err
}

// DefaultFuncs returns the standard template functions for Datastar.
func DefaultFuncs() template.FuncMap {
	return template.FuncMap{
		// Datastar action helpers (data-on:event="@method('/url')")
		"dsGet":    dsGet,
		"dsPost":   dsPost,
		"dsPut":    dsPut,
		"dsPatch":  dsPatch,
		"dsDelete": dsDelete,

		// Datastar event handlers
		"dsOnClick":   dsOnClick,
		"dsOnSubmit":  dsOnSubmit,
		"dsOnChange":  dsOnChange,
		"dsOnInput":   dsOnInput,
		"dsOnKeyup":   dsOnKeyup,
		"dsOnLoad":    dsOnLoad,
		"dsOnIntersect": dsOnIntersect,

		// Datastar binding and signals
		"dsBind":     dsBind,
		"dsSignals":  dsSignals,
		"dsSignalsJSON": dsSignalsJSON,

		// Datastar display helpers
		"dsText":  dsText,
		"dsShow":  dsShow,
		"dsClass": dsClass,
		"dsAttr":  dsAttr,
		"dsStyle": dsStyle,

		// Datastar indicators and refs
		"dsIndicator": dsIndicator,
		"dsRef":       dsRef,

		// HTML helpers
		"safe":    safe,
		"safeURL": safeURL,
		"attr":    attr,
		"class":   class,

		// Utility helpers
		"join":      strings.Join,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"trim":      strings.TrimSpace,
		"replace":   strings.ReplaceAll,

		// Conditional helpers
		"if_":     ifFunc,
		"default": defaultFunc,
		"coalesce": coalesce,
	}
}

// --- Datastar Action Helpers ---
// These generate data-on:click="@method('/url')" style attributes

// dsGet generates a data-on:click attribute with @get action
func dsGet(url string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:click="@get('` + url + `')"`)
}

// dsPost generates a data-on:click attribute with @post action
func dsPost(url string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:click="@post('` + url + `')"`)
}

// dsPut generates a data-on:click attribute with @put action
func dsPut(url string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:click="@put('` + url + `')"`)
}

// dsPatch generates a data-on:click attribute with @patch action
func dsPatch(url string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:click="@patch('` + url + `')"`)
}

// dsDelete generates a data-on:click attribute with @delete action
func dsDelete(url string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:click="@delete('` + url + `')"`)
}

// --- Datastar Event Handlers ---

// dsOnClick generates a data-on:click attribute with custom expression
func dsOnClick(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:click="` + expression + `"`)
}

// dsOnSubmit generates a data-on:submit attribute
func dsOnSubmit(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:submit="` + expression + `"`)
}

// dsOnChange generates a data-on:change attribute
func dsOnChange(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:change="` + expression + `"`)
}

// dsOnInput generates a data-on:input attribute
func dsOnInput(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:input="` + expression + `"`)
}

// dsOnKeyup generates a data-on:keyup attribute
func dsOnKeyup(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:keyup="` + expression + `"`)
}

// dsOnLoad generates a data-on:load attribute (triggers when element loads)
func dsOnLoad(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-on:load="` + expression + `"`)
}

// dsOnIntersect generates a data-on-intersect attribute (triggers when visible)
func dsOnIntersect(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-on-intersect="` + expression + `"`)
}

// --- Datastar Binding and Signals ---

// dsBind generates a data-bind:signal attribute for two-way binding
func dsBind(signal string) template.HTMLAttr {
	return template.HTMLAttr(`data-bind:` + signal)
}

// dsSignals generates a data-signals attribute with raw JSON
func dsSignals(json string) template.HTMLAttr {
	return template.HTMLAttr(`data-signals="` + json + `"`)
}

// dsSignalsJSON generates a data-signals attribute from a Go map
func dsSignalsJSON(data any) template.HTMLAttr {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return template.HTMLAttr(`data-signals="` + string(jsonBytes) + `"`)
}

// --- Datastar Display Helpers ---

// dsText generates a data-text attribute for reactive text content
func dsText(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-text="` + expression + `"`)
}

// dsShow generates a data-show attribute for conditional visibility
func dsShow(expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-show="` + expression + `"`)
}

// dsClass generates a data-class:classname attribute for conditional classes
func dsClass(className, expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-class:` + className + `="` + expression + `"`)
}

// dsAttr generates a data-attr:attrname attribute for reactive attributes
func dsAttr(attrName, expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-attr:` + attrName + `="` + expression + `"`)
}

// dsStyle generates a data-style:property attribute for reactive inline styles
func dsStyle(property, expression string) template.HTMLAttr {
	return template.HTMLAttr(`data-style:` + property + `="` + expression + `"`)
}

// --- Datastar Indicators and Refs ---

// dsIndicator generates a data-indicator:signal attribute for loading states
func dsIndicator(signal string) template.HTMLAttr {
	return template.HTMLAttr(`data-indicator:` + signal)
}

// dsRef generates a data-ref:name attribute for element references
func dsRef(name string) template.HTMLAttr {
	return template.HTMLAttr(`data-ref:` + name)
}

// --- HTML Helpers ---

func safe(s string) template.HTML {
	return template.HTML(s)
}

func safeURL(s string) template.URL {
	return template.URL(s)
}

func attr(key, value string) template.HTMLAttr {
	return template.HTMLAttr(key + `="` + value + `"`)
}

func class(classes ...string) template.HTMLAttr {
	var nonEmpty []string
	for _, c := range classes {
		if c = strings.TrimSpace(c); c != "" {
			nonEmpty = append(nonEmpty, c)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	return template.HTMLAttr(`class="` + strings.Join(nonEmpty, " ") + `"`)
}

// --- Conditional Helpers ---

func ifFunc(condition bool, trueVal, falseVal any) any {
	if condition {
		return trueVal
	}
	return falseVal
}

func defaultFunc(val, def any) any {
	if val == nil || val == "" || val == 0 || val == false {
		return def
	}
	return val
}

func coalesce(vals ...any) any {
	for _, v := range vals {
		if v != nil && v != "" && v != 0 && v != false {
			return v
		}
	}
	return nil
}
