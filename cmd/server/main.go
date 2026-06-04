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
	"github.com/vesper/snip/internal/middleware"
	"github.com/vesper/snip/internal/services"
	"github.com/vesper/snip/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := store.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	pasteSvc := services.NewPasteService(db, cfg.Paste.MaxSize, cfg.Server.BaseURL)
	h := handlers.New(pasteSvc, db, cfg)

	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.RealIP)
	r.Use(middleware.WithClientIP)
	r.Use(middleware.Security)
	r.Use(chimw.Compress(5))
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(middleware.RateLimit(120, time.Minute))

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", handlers.StaticHandler()))

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
	})

	// HTMX
	r.Post("/hx/paste", h.CreateHTMX)

	// API
	r.Route("/api/v1", func(r chi.Router) {
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
	go func() {
		for range time.Tick(10 * time.Minute) {
			if n, _ := pasteSvc.Cleanup(); n > 0 {
				log.Printf("Cleaned %d expired pastes", n)
			}
		}
	}()

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Snip running at %s", cfg.Server.BaseURL)

	srv := &http.Server{Addr: addr, Handler: r, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		log.Println("Shutting down...")
		srv.Close()
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
