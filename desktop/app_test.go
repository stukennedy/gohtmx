package desktop

import (
	"net/http"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Title != "Irgo App" {
		t.Errorf("expected default title 'Irgo App', got %q", config.Title)
	}
	if config.Width != 1024 {
		t.Errorf("expected default width 1024, got %d", config.Width)
	}
	if config.Height != 768 {
		t.Errorf("expected default height 768, got %d", config.Height)
	}
	if !config.Resizable {
		t.Error("expected default Resizable to be true")
	}
	if config.Debug {
		t.Error("expected default Debug to be false")
	}
	if config.Port != 0 {
		t.Errorf("expected default Port 0 (auto), got %d", config.Port)
	}
	if config.Transport != "loopback" {
		t.Errorf("expected default Transport 'loopback', got %q", config.Transport)
	}
}

func TestNewApp(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	config := DefaultConfig()
	app := New(handler, config)

	if app == nil {
		t.Fatal("expected non-nil app")
	}
	if app.handler == nil {
		t.Error("expected handler to be set")
	}
	if app.config.Title != config.Title {
		t.Error("expected config to be set")
	}
	if app.wsHub == nil {
		t.Error("expected wsHub to be initialized")
	}
}

func TestNewAppWithHub(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	config := DefaultConfig()

	// Create app with custom hub (nil for this test - just testing the API)
	app := NewWithHub(handler, nil, config)

	if app == nil {
		t.Fatal("expected non-nil app")
	}
	if app.wsHub != nil {
		t.Error("expected nil wsHub when nil passed to NewWithHub")
	}
}

func TestAppURLBeforeRun(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	config := DefaultConfig()
	app := New(handler, config)

	// Before Run(), URL should be empty since transport isn't started
	url := app.URL()
	if url != "" {
		t.Errorf("expected empty URL before Run(), got %q", url)
	}
}

func TestAppPortBeforeRun(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	config := DefaultConfig()
	app := New(handler, config)

	// Before Run(), Port should be 0 since transport isn't started
	port := app.Port()
	if port != 0 {
		t.Errorf("expected port 0 before Run(), got %d", port)
	}
}

func TestAppSecretBeforeRun(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	config := DefaultConfig()
	app := New(handler, config)

	// Before Run(), Secret should be empty since transport isn't started
	secret := app.Secret()
	if secret != "" {
		t.Errorf("expected empty secret before Run(), got %q", secret)
	}
}

func TestAppHub(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	config := DefaultConfig()
	app := New(handler, config)

	hub := app.Hub()
	if hub == nil {
		t.Error("expected Hub() to return non-nil hub")
	}
}

func TestGenerateSecret(t *testing.T) {
	secret1, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}
	if len(secret1) == 0 {
		t.Error("expected non-empty secret")
	}

	// Should be URL-safe base64 encoded 32 bytes = 44 chars (with padding)
	if len(secret1) < 40 {
		t.Errorf("expected secret of at least 40 chars, got %d chars", len(secret1))
	}

	// Secrets should be unique
	secret2, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret failed: %v", err)
	}
	if secret1 == secret2 {
		t.Error("expected unique secrets on each call")
	}
}
