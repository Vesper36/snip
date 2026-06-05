package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vesper/snip/internal/auth"
	"github.com/vesper/snip/internal/store"
)

type ctxKey string

const (
	CtxClaims   ctxKey = "claims"
	CtxAPIToken ctxKey = "api_token"
	CtxClientIP ctxKey = "client_ip"
	CtxLang     ctxKey = "lang"
)

func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, _ := strings.Cut(r.RemoteAddr, ":")
	return host
}

func WithClientIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), CtxClientIP, ClientIP(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// CSP allows inline styles for our gradient text + CDN scripts
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://unpkg.com ; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; img-src 'self' data: https:; font-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

func RateLimit(max int, window time.Duration) func(http.Handler) http.Handler {
	type entry struct {
		n   int
		at  time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*entry)
	)
	go func() {
		for {
			time.Sleep(window)
			mu.Lock()
			for k, v := range clients {
				if time.Since(v.at) > window {
					delete(clients, k)
				}
			}
			mu.Unlock()
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ClientIP(r)
			mu.Lock()
			e, ok := clients[ip]
			if !ok || time.Since(e.at) > window {
				clients[ip] = &entry{1, time.Now()}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}
			if e.n >= max {
				mu.Unlock()
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			e.n++
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var raw string
			if h := r.Header.Get("Authorization"); h != "" {
				if p := strings.SplitN(h, " ", 2); len(p) == 2 {
					raw = p[1]
				}
			}
			if raw == "" {
				if c, err := r.Cookie("snip_token"); err == nil {
					raw = c.Value
				}
			}
			if raw == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, err := auth.ValidateJWT(secret, raw)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), CtxClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func APITokenAuth(s *store.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var raw string
			if h := r.Header.Get("Authorization"); h != "" {
				if p := strings.SplitN(h, " ", 2); len(p) == 2 {
					raw = p[1]
				}
			}
			if raw == "" {
				http.Error(w, "api token required", http.StatusUnauthorized)
				return
			}
			t, err := s.GetTokenByHash(store.HashToken(raw))
			if err != nil || t == nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
				http.Error(w, "token expired", http.StatusUnauthorized)
				return
			}
			s.UpdateTokenUsed(t.ID)
			ctx := context.WithValue(r.Context(), CtxAPIToken, t)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = 200
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// RequestLogger logs incoming requests.
func RequestLogger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, statusCode: 200}
			next.ServeHTTP(rw, r)
			if rw.statusCode >= 400 || r.URL.Path == "/healthz" || r.URL.Path == "/metrics" {
				return // skip noisy or error logs
			}
			log.Printf("%s %s %s %d %s",
				r.Method, r.URL.Path, r.RemoteAddr,
				rw.statusCode, time.Since(start).Round(time.Millisecond))
		})
	}
}

// CORS adds CORS headers for API routes.
func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// DetectLanguage sets the language in context from cookie or Accept-Language header.
func DetectLanguage(supported []string, defaultLang string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := defaultLang

			// 1. Check cookie
			if c, err := r.Cookie("lang"); err == nil {
				for _, s := range supported {
					if c.Value == s {
						lang = c.Value
						break
					}
				}
			} else if al := r.Header.Get("Accept-Language"); al != "" {
				// 2. Parse Accept-Language
				for _, part := range strings.Split(al, ",") {
					code := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
					code = strings.SplitN(code, "-", 2)[0]
					for _, s := range supported {
						if code == s {
							lang = code
							break
						}
					}
					if lang != defaultLang {
						break
					}
				}
			}

			ctx := context.WithValue(r.Context(), CtxLang, lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
