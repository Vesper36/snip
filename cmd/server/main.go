package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/vesper/snip/internal/config"
	"github.com/vesper/snip/internal/handlers"
	"github.com/vesper/snip/internal/i18n"
	"github.com/vesper/snip/internal/middleware"
	"github.com/vesper/snip/internal/services"
	"github.com/vesper/snip/internal/store"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cfg := config.Load()

	db, err := store.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	pasteSvc := services.NewPasteService(db, cfg.Paste.MaxSize, cfg.Server.BaseURL)
	h := handlers.New(pasteSvc, db, cfg, version)

	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.RealIP)
	r.Use(middleware.WithClientIP)
	r.Use(middleware.Security)
	r.Use(middleware.RequestLogger())
	r.Use(chimw.Compress(5))
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(middleware.RateLimit(120, time.Minute))
	r.Use(middleware.DetectLanguage(i18n.Supported(), i18n.DefaultLang))

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", handlers.StaticHandler()))

	// Language switcher
	r.Get("/lang", h.SetLang)
	r.Post("/lang", h.SetLang)

	// Web routes
	r.Get("/", h.Home)
	r.Get("/search", h.Search)
	r.Get("/my", h.List)
	r.Get("/settings", h.Settings)

	// Paste routes
	r.Route("/{slug}", func(r chi.Router) {
		r.Get("/", h.View)
		r.Get("/raw", h.Raw)
		r.Get("/download", h.Download)
		r.Post("/fork", h.Fork)
		r.Get("/fork", h.Fork)
	})

	// HTMX
	r.Post("/hx/paste", h.CreateHTMX)

	// Health & metrics
	r.Get("/healthz", h.Healthz)
	r.Get("/metrics", h.Metrics)

	// API
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.CORS())
		r.Post("/pastes", h.APICreate)
		r.Get("/pastes", h.APIList)
		r.Get("/pastes/{slug}", h.APIGet)
		r.Get("/search", h.APISearch)
		r.Get("/stats", h.APIStats)

		r.Group(func(r chi.Router) {
			r.Use(middleware.APITokenAuth(db))
			r.Delete("/pastes/{slug}", h.APIDelete)
			r.Post("/tokens", h.APICreateToken)
			r.Delete("/tokens/{id}", h.APIDeleteToken)
		})
	})

	// Cleanup expired pastes
	stopCleanup := make(chan struct{})
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if n, _ := pasteSvc.Cleanup(); n > 0 {
					log.Printf("Cleaned %d expired pastes", n)
				}
			case <-stopCleanup:
				return
			}
		}
	}()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Snip running at %s", cfg.Server.BaseURL)

	srv := &http.Server{Addr: addr, Handler: r, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down...")
		close(stopCleanup)
		srv.Close()
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
