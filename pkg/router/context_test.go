package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextParam(t *testing.T) {
	r := New()
	var capturedID string

	r.GET("/users/{id}", func(ctx *Context) (string, error) {
		capturedID = ctx.Param("id")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if capturedID != "123" {
		t.Errorf("expected param id='123', got %q", capturedID)
	}
}

func TestContextQuery(t *testing.T) {
	r := New()
	var capturedQ, capturedPage string

	r.GET("/search", func(ctx *Context) (string, error) {
		capturedQ = ctx.Query("q")
		capturedPage = ctx.Query("page")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/search?q=hello&page=2", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if capturedQ != "hello" {
		t.Errorf("expected query q='hello', got %q", capturedQ)
	}
	if capturedPage != "2" {
		t.Errorf("expected query page='2', got %q", capturedPage)
	}
}

func TestContextQueryDefault(t *testing.T) {
	r := New()
	var capturedPage string

	r.GET("/search", func(ctx *Context) (string, error) {
		capturedPage = ctx.QueryDefault("page", "1")
		return "", nil
	})

	// Test with no page param
	req := httptest.NewRequest("GET", "/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if capturedPage != "1" {
		t.Errorf("expected default page='1', got %q", capturedPage)
	}

	// Test with page param
	req = httptest.NewRequest("GET", "/search?page=5", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if capturedPage != "5" {
		t.Errorf("expected page='5', got %q", capturedPage)
	}
}

func TestContextFormValue(t *testing.T) {
	r := New()
	var capturedName string

	r.POST("/submit", func(ctx *Context) (string, error) {
		capturedName = ctx.FormValue("name")
		return "", nil
	})

	body := strings.NewReader("name=John")
	req := httptest.NewRequest("POST", "/submit", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if capturedName != "John" {
		t.Errorf("expected form name='John', got %q", capturedName)
	}
}

func TestContextIsHTMX(t *testing.T) {
	r := New()
	var isHTMX bool

	r.GET("/test", func(ctx *Context) (string, error) {
		isHTMX = ctx.IsHTMX()
		return "", nil
	})

	// Test non-HTMX request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if isHTMX {
		t.Error("expected IsHTMX()=false for regular request")
	}

	// Test HTMX request
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("HX-Request", "true")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !isHTMX {
		t.Error("expected IsHTMX()=true for HTMX request")
	}
}

func TestContextHTMXHeaders(t *testing.T) {
	r := New()
	var target, trigger, triggerName, currentURL, prompt string

	r.GET("/test", func(ctx *Context) (string, error) {
		target = ctx.HXTarget()
		trigger = ctx.HXTrigger()
		triggerName = ctx.HXTriggerName()
		currentURL = ctx.HXCurrentURL()
		prompt = ctx.HXPrompt()
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("HX-Target", "#content")
	req.Header.Set("HX-Trigger", "btn-1")
	req.Header.Set("HX-Trigger-Name", "submit")
	req.Header.Set("HX-Current-URL", "http://example.com/page")
	req.Header.Set("HX-Prompt", "user input")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if target != "#content" {
		t.Errorf("expected HXTarget='#content', got %q", target)
	}
	if trigger != "btn-1" {
		t.Errorf("expected HXTrigger='btn-1', got %q", trigger)
	}
	if triggerName != "submit" {
		t.Errorf("expected HXTriggerName='submit', got %q", triggerName)
	}
	if currentURL != "http://example.com/page" {
		t.Errorf("expected HXCurrentURL='http://example.com/page', got %q", currentURL)
	}
	if prompt != "user input" {
		t.Errorf("expected HXPrompt='user input', got %q", prompt)
	}
}

func TestContextHeader(t *testing.T) {
	r := New()
	var customHeader string

	r.GET("/test", func(ctx *Context) (string, error) {
		customHeader = ctx.Header("X-Custom")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Custom", "custom-value")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if customHeader != "custom-value" {
		t.Errorf("expected header X-Custom='custom-value', got %q", customHeader)
	}
}

func TestContextSetHeader(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.SetHeader("X-Response", "response-value")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("X-Response") != "response-value" {
		t.Errorf("expected response header X-Response='response-value', got %q", w.Header().Get("X-Response"))
	}
}

func TestContextHTML(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.HTML("<div>Direct HTML</div>")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "<div>Direct HTML</div>" {
		t.Errorf("expected '<div>Direct HTML</div>', got %q", w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got %q", ct)
	}
}

func TestContextHTMLStatus(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.HTMLStatus(http.StatusCreated, "<div>Created</div>")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestContextJSON(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.JSON(map[string]string{"status": "ok"})
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("expected JSON with status:ok, got %q", w.Body.String())
	}
}

func TestContextError(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.ErrorStatus(http.StatusBadRequest, "Invalid input")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Invalid input") {
		t.Errorf("expected error message 'Invalid input', got %q", w.Body.String())
	}
}

func TestContextNotFound(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.NotFound("Item not found")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestContextBadRequest(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.BadRequest("Missing required field")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestContextTrigger(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.Trigger("itemCreated")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Trigger") != "itemCreated" {
		t.Errorf("expected HX-Trigger='itemCreated', got %q", w.Header().Get("HX-Trigger"))
	}
}

func TestContextTriggerAfterSettle(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.TriggerAfterSettle("animationComplete")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Trigger-After-Settle") != "animationComplete" {
		t.Errorf("expected HX-Trigger-After-Settle='animationComplete', got %q", w.Header().Get("HX-Trigger-After-Settle"))
	}
}

func TestContextRedirect(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.Redirect("/new-location")
		return "", nil
	})

	// Test regular redirect
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}

	// Test HTMX redirect
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("HX-Request", "true")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Redirect") != "/new-location" {
		t.Errorf("expected HX-Redirect='/new-location', got %q", w.Header().Get("HX-Redirect"))
	}
}

func TestContextPushURL(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.PushURL("/new-url")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Push-Url") != "/new-url" {
		t.Errorf("expected HX-Push-Url='/new-url', got %q", w.Header().Get("HX-Push-Url"))
	}
}

func TestContextReplaceURL(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.ReplaceURL("/replaced-url")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Replace-Url") != "/replaced-url" {
		t.Errorf("expected HX-Replace-Url='/replaced-url', got %q", w.Header().Get("HX-Replace-Url"))
	}
}

func TestContextRetarget(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.Retarget("#new-target")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Retarget") != "#new-target" {
		t.Errorf("expected HX-Retarget='#new-target', got %q", w.Header().Get("HX-Retarget"))
	}
}

func TestContextReswap(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.Reswap("outerHTML")
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Reswap") != "outerHTML" {
		t.Errorf("expected HX-Reswap='outerHTML', got %q", w.Header().Get("HX-Reswap"))
	}
}

func TestContextRefresh(t *testing.T) {
	r := New()

	r.GET("/test", func(ctx *Context) (string, error) {
		ctx.Refresh()
		return "", nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("HX-Refresh") != "true" {
		t.Errorf("expected HX-Refresh='true', got %q", w.Header().Get("HX-Refresh"))
	}
}

func TestContextNoContent(t *testing.T) {
	r := New()

	r.DELETE("/test", func(ctx *Context) (string, error) {
		ctx.NoContent()
		return "", nil
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

func TestContextBind(t *testing.T) {
	r := New()

	type Input struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	var input Input

	r.POST("/test", func(ctx *Context) (string, error) {
		if err := ctx.Bind(&input); err != nil {
			ctx.BadRequest(err.Error())
			return "", nil
		}
		return "", nil
	})

	body := strings.NewReader(`{"name":"John","age":30}`)
	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if input.Name != "John" {
		t.Errorf("expected name='John', got %q", input.Name)
	}
	if input.Age != 30 {
		t.Errorf("expected age=30, got %d", input.Age)
	}
}

func TestContextWritten(t *testing.T) {
	ctx := NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	if ctx.Written() {
		t.Error("expected Written()=false before writing")
	}

	ctx.HTML("<div>test</div>")

	if !ctx.Written() {
		t.Error("expected Written()=true after writing")
	}
}
