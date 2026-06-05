package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIP_XForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	ip := ClientIP(r)
	if ip != "1.2.3.4" {
		t.Fatalf("Expected 1.2.3.4, got %s", ip)
	}
}

func TestClientIP_XRealIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Real-IP", "9.8.7.6")
	ip := ClientIP(r)
	if ip != "9.8.7.6" {
		t.Fatalf("Expected 9.8.7.6, got %s", ip)
	}
}

func TestClientIP_RemoteAddr(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.168.1.1:12345"
	ip := ClientIP(r)
	if ip != "192.168.1.1" {
		t.Fatalf("Expected 192.168.1.1, got %s", ip)
	}
}

func TestSecurity_Headers(t *testing.T) {
	handler := Security(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)

	tests := []struct {
		header   string
		expected string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}
	for _, tt := range tests {
		got := w.Header().Get(tt.header)
		if got != tt.expected {
			t.Errorf("Header %s = %q, want %q", tt.header, got, tt.expected)
		}
	}
	csp := w.Header().Get("Content-Security-Policy")
	if csp == "" {
		t.Error("CSP header is empty")
	}
}

func TestRateLimit_AllowsNormal(t *testing.T) {
	handler := RateLimit(10, 60)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	for i := 0; i < 10; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "10.0.0.1:1234"
		handler.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("Request %d: expected 200, got %d", i, w.Code)
		}
	}
}

func TestRateLimit_BlocksExcess(t *testing.T) {
	// Use a short window to avoid timing issues
	handler := RateLimit(2, 60*time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	// Make 2 requests to exhaust limit
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Real-IP", "10.0.0.88")
		handler.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("Request %d: expected 200, got %d", i, w.Code)
		}
	}
	// Third request should be blocked
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Real-IP", "10.0.0.88")
	handler.ServeHTTP(w, r)
	if w.Code != 429 {
		t.Fatalf("Expected 429, got %d", w.Code)
	}
}

func TestCORS_Headers(t *testing.T) {
	handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Missing CORS Allow-Origin header")
	}
}

func TestCORS_OptionsRequest(t *testing.T) {
	handler := CORS()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/", nil)
	handler.ServeHTTP(w, r)
	if w.Code != 204 {
		t.Fatalf("Expected 204 for OPTIONS, got %d", w.Code)
	}
}

func TestDetectLanguage_Cookie(t *testing.T) {
	handler := DetectLanguage([]string{"en", "zh", "ja"}, "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Context().Value(CtxLang).(string)
		w.Write([]byte(lang))
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "lang", Value: "zh"})
	handler.ServeHTTP(w, r)
	if w.Body.String() != "zh" {
		t.Fatalf("Expected zh from cookie, got %s", w.Body.String())
	}
}

func TestDetectLanguage_AcceptLanguage(t *testing.T) {
	handler := DetectLanguage([]string{"en", "zh", "ja"}, "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Context().Value(CtxLang).(string)
		w.Write([]byte(lang))
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Language", "ja-JP,ja;q=0.9")
	handler.ServeHTTP(w, r)
	if w.Body.String() != "ja" {
		t.Fatalf("Expected ja from Accept-Language, got %s", w.Body.String())
	}
}

func TestDetectLanguage_Default(t *testing.T) {
	handler := DetectLanguage([]string{"en", "zh"}, "en")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := r.Context().Value(CtxLang).(string)
		w.Write([]byte(lang))
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	if w.Body.String() != "en" {
		t.Fatalf("Expected en as default, got %s", w.Body.String())
	}
}

func TestRequestLogger(t *testing.T) {
	handler := RequestLogger()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("Expected 200, got %d", w.Code)
	}
}
