package render

import (
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

// DefaultFuncs returns the standard template functions for HTMX.
func DefaultFuncs() template.FuncMap {
	return template.FuncMap{
		// HTMX attribute helpers
		"hxGet":     hxGet,
		"hxPost":    hxPost,
		"hxPut":     hxPut,
		"hxPatch":   hxPatch,
		"hxDelete":  hxDelete,
		"hxTarget":  hxTarget,
		"hxSwap":    hxSwap,
		"hxTrigger": hxTrigger,
		"hxVals":    hxVals,
		"hxConfirm": hxConfirm,
		"hxBoost":   hxBoost,
		"hxPushURL": hxPushURL,
		"hxInclude": hxInclude,
		"hxSelect":  hxSelect,
		"hxExt":     hxExt,

		// WebSocket helpers (HTMX 4)
		"hxWsConnect": hxWsConnect,
		"hxWsSend":    hxWsSend,

		// HTML helpers
		"safe":    safe,
		"safeURL": safeURL,
		"attr":    attr,
		"class":   class,

		// Utility helpers
		"join":     strings.Join,
		"contains": strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"lower":    strings.ToLower,
		"upper":    strings.ToUpper,
		"trim":     strings.TrimSpace,
		"replace":  strings.ReplaceAll,

		// Conditional helpers
		"if_":     ifFunc,
		"default": defaultFunc,
		"coalesce": coalesce,
	}
}

// HTMX attribute functions
func hxGet(url string) template.HTMLAttr {
	return template.HTMLAttr(`hx-get="` + url + `"`)
}

func hxPost(url string) template.HTMLAttr {
	return template.HTMLAttr(`hx-post="` + url + `"`)
}

func hxPut(url string) template.HTMLAttr {
	return template.HTMLAttr(`hx-put="` + url + `"`)
}

func hxPatch(url string) template.HTMLAttr {
	return template.HTMLAttr(`hx-patch="` + url + `"`)
}

func hxDelete(url string) template.HTMLAttr {
	return template.HTMLAttr(`hx-delete="` + url + `"`)
}

func hxTarget(selector string) template.HTMLAttr {
	return template.HTMLAttr(`hx-target="` + selector + `"`)
}

func hxSwap(strategy string) template.HTMLAttr {
	return template.HTMLAttr(`hx-swap="` + strategy + `"`)
}

func hxTrigger(trigger string) template.HTMLAttr {
	return template.HTMLAttr(`hx-trigger="` + trigger + `"`)
}

func hxVals(json string) template.HTMLAttr {
	return template.HTMLAttr(`hx-vals='` + json + `'`)
}

func hxConfirm(message string) template.HTMLAttr {
	return template.HTMLAttr(`hx-confirm="` + message + `"`)
}

func hxBoost(enabled bool) template.HTMLAttr {
	if enabled {
		return template.HTMLAttr(`hx-boost="true"`)
	}
	return template.HTMLAttr(`hx-boost="false"`)
}

func hxPushURL(url string) template.HTMLAttr {
	if url == "true" || url == "false" {
		return template.HTMLAttr(`hx-push-url="` + url + `"`)
	}
	return template.HTMLAttr(`hx-push-url="` + url + `"`)
}

func hxInclude(selector string) template.HTMLAttr {
	return template.HTMLAttr(`hx-include="` + selector + `"`)
}

func hxSelect(selector string) template.HTMLAttr {
	return template.HTMLAttr(`hx-select="` + selector + `"`)
}

func hxExt(extension string) template.HTMLAttr {
	return template.HTMLAttr(`hx-ext="` + extension + `"`)
}

// WebSocket helpers (HTMX 4)
func hxWsConnect(url string) template.HTMLAttr {
	return template.HTMLAttr(`hx-ws:connect="` + url + `"`)
}

func hxWsSend(val string) template.HTMLAttr {
	if val == "" {
		return template.HTMLAttr(`hx-ws:send`)
	}
	return template.HTMLAttr(`hx-ws:send="` + val + `"`)
}

// HTML helpers
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

// Conditional helpers
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
