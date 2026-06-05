package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vesper/snip/internal/config"
	"github.com/vesper/snip/internal/services"
	"github.com/vesper/snip/internal/store"
)

func setupTestHandler(t *testing.T) (*Handler, func()) {
	t.Helper()
	db, err := store.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test db: %v", err)
	}
	cfg := &config.Config{
		Server: config.ServerConfig{BaseURL: "http://localhost:53524"},
		Paste:  config.PasteConfig{MaxSize: 1024 * 1024},
	}
	ps := services.NewPasteService(db, cfg.Paste.MaxSize, cfg.Server.BaseURL)
	h := New(ps, db, cfg, "test")
	return h, func() { db.Close() }
}

func TestHomeHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	h.Home(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if len(body) == 0 {
		t.Fatal("Empty response body")
	}
}

func TestHealthzHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/healthz", nil)
	h.Healthz(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestMetricsHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/metrics", nil)
	h.Metrics(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if len(body) == 0 {
		t.Fatal("Empty metrics body")
	}
}

func TestSearchHandler_RedirectOnEmpty(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search", nil)
	h.Search(w, r)
	if w.Code != 302 {
		t.Fatalf("Expected 302 redirect, got %d", w.Code)
	}
}

func TestViewHandler_NotFound(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/nonexistent", nil)
	r = addLangContext(r)
	h.View(w, r)
	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestRawHandler_NotFound(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/nonexistent/raw", nil)
	h.Raw(w, r)
	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestDownloadHandler_NotFound(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/nonexistent/download", nil)
	h.Download(w, r)
	if w.Code != 404 {
		t.Fatalf("Expected 404, got %d", w.Code)
	}
}

func TestAPICreate_EmptyContent(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/pastes", nil)
	r.Header.Set("Content-Type", "application/json")
	h.APICreate(w, r)
	if w.Code != 400 {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func TestAPICreate_InvalidJSON(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/pastes", nil)
	r.Header.Set("Content-Type", "application/json")
	h.APICreate(w, r)
	if w.Code != 400 {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func TestAPIListHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/pastes", nil)
	h.APIList(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestAPISearchHandler_EmptyQuery(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/search", nil)
	h.APISearch(w, r)
	if w.Code != 400 {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func TestAPIStatsHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/stats", nil)
	h.APIStats(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestSetLangHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/lang?lang=zh", nil)
	h.SetLang(w, r)
	if w.Code != 303 {
		t.Fatalf("Expected 303 redirect, got %d", w.Code)
	}
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "lang" && c.Value == "zh" {
			found = true
		}
	}
	if !found {
		t.Fatal("Expected lang cookie set to zh")
	}
}

func TestSetLangHandler_InvalidLang(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/lang?lang=xx", nil)
	h.SetLang(w, r)
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "lang" && c.Value != "en" {
			t.Fatalf("Expected fallback to en, got %s", c.Value)
		}
	}
}

func TestListHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/my", nil)
	r = addLangContext(r)
	h.List(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func TestSettingsHandler(t *testing.T) {
	h, cleanup := setupTestHandler(t)
	defer cleanup()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/settings", nil)
	r = addLangContext(r)
	h.Settings(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}

func addLangContext(r *http.Request) *http.Request {
	ctx := r.Context()
	// Use the middleware's context key
	type ctxKey string
	return r.WithContext(ctx)
}
