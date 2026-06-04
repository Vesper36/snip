package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/vesper/snip/internal/auth"
	"github.com/vesper/snip/internal/config"
	"github.com/vesper/snip/internal/middleware"
	"github.com/vesper/snip/internal/models"
	"github.com/vesper/snip/internal/services"
	"github.com/vesper/snip/internal/store"
)

type Handler struct {
	paste    *services.PasteService
	store    *store.Store
	cfg      *config.Config
	tmpls    map[string]*template.Template
}

func New(ps *services.PasteService, s *store.Store, cfg *config.Config) *Handler {
	funcMap := template.FuncMap{
		"size": func(b int64) string {
			switch {
			case b < 1024:
				return fmt.Sprintf("%d B", b)
			case b < 1024*1024:
				return fmt.Sprintf("%.1f KB", float64(b)/1024)
			default:
				return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
			}
		},
		"timeAgo": func(t time.Time) string {
			d := time.Since(t)
			switch {
			case d < time.Minute:
				return "just now"
			case d < time.Hour:
				return fmt.Sprintf("%dm ago", int(d.Minutes()))
			case d < 24*time.Hour:
				return fmt.Sprintf("%dh ago", int(d.Hours()))
			default:
				return fmt.Sprintf("%dd ago", int(d.Hours()/24))
			}
		},
		"langClass": func(l string) string { return "language-" + l },
		"safe":      func(s string) template.HTML { return template.HTML(s) },
		"add":       func(a, b int) int { return a + b },
		"sub":       func(a, b int) int { return a - b },
	}
	tmpls := parseTemplates(funcMap)
	return &Handler{paste: ps, store: s, cfg: cfg, tmpls: tmpls}
}

func (h *Handler) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// name is like "home.html", template key is "home"
	key := name
	if idx := len(key) - 5; idx > 0 && key[idx:] == ".html" {
		key = key[:idx]
	}
	t, ok := h.tmpls[key]
	if !ok {
		http.Error(w, "template not found: "+name, 500)
		return
	}
	t.ExecuteTemplate(w, name, data)
}

func (h *Handler) errPage(w http.ResponseWriter, code int, title, msg string) {
	w.WriteHeader(code)
	h.render(w, "error.html", map[string]any{"Code": code, "Title": title, "Message": msg})
}

// --- Web Handlers ---

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, "home.html", map[string]any{
		"Languages": models.LanguageList(),
		"BaseURL":   h.cfg.Server.BaseURL,
	})
}

func (h *Handler) View(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	pw := r.URL.Query().Get("password")
	if pw == "" {
		pw = r.FormValue("password")
	}

	p, err := h.paste.Get(slug, pw)
	if err != nil {
		switch err {
		case services.ErrNotFound:
			h.errPage(w, 404, "Not Found", "This paste does not exist.")
		case services.ErrExpired:
			h.errPage(w, 410, "Expired", "This paste has expired.")
		case services.ErrMaxViews:
			h.errPage(w, 410, "Gone", "View limit reached.")
		case services.ErrNeedPassword:
			h.render(w, "password.html", map[string]any{"Slug": slug, "NeedPW": true})
		case services.ErrBadPassword:
			h.render(w, "password.html", map[string]any{"Slug": slug, "NeedPW": true, "Error": "Wrong password."})
		default:
			h.errPage(w, 500, "Error", "Something went wrong.")
		}
		return
	}
	h.render(w, "view.html", map[string]any{"P": p, "BaseURL": h.cfg.Server.BaseURL, "Languages": models.LanguageList()})
}

func (h *Handler) Raw(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.paste.Get(slug, r.URL.Query().Get("password"))
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(p.Content))
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.paste.Get(slug, r.URL.Query().Get("password"))
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}
	name := p.Title
	if name == "" {
		name = slug + ".txt"
	}
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write([]byte(p.Content))
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	pastes, _ := h.paste.Search(q, 50)
	h.render(w, "search.html", map[string]any{"Query": q, "Pastes": pastes, "BaseURL": h.cfg.Server.BaseURL})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	pastes, _ := h.paste.List(limit+1, (page-1)*limit)
	hasNext := len(pastes) > limit
	if hasNext {
		pastes = pastes[:limit]
	}
	total, _ := h.store.CountPastes()
	h.render(w, "list.html", map[string]any{
		"Pastes": pastes, "Page": page, "HasNext": hasNext, "Total": total, "BaseURL": h.cfg.Server.BaseURL,
	})
}

func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	tokens, _ := h.store.ListTokens()
	stats, _ := h.paste.Stats()
	h.render(w, "settings.html", map[string]any{"Tokens": tokens, "Stats": stats, "BaseURL": h.cfg.Server.BaseURL})
}

// --- HTMX Handler ---

func (h *Handler) CreateHTMX(w http.ResponseWriter, r *http.Request) {
	req := models.PasteCreateRequest{
		Title:    r.FormValue("title"),
		Content:  r.FormValue("content"),
		Language: r.FormValue("language"),
		Password: r.FormValue("password"),
		ExpiresIn: r.FormValue("expires_in"),
		CustomSlug: r.FormValue("custom_slug"),
	}
	if r.FormValue("burn_after_read") == "on" {
		req.BurnAfterRead = true
	}
	if v := r.FormValue("max_views"); v != "" {
		req.MaxViews, _ = strconv.Atoi(v)
	}

	ip := ""
	if v := r.Context().Value(middleware.CtxClientIP); v != nil {
		ip = v.(string)
	}

	p, err := h.paste.Create(req, ip)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`<div class="alert alert-err">` + err.Error() + `</div>`))
		return
	}
	w.Header().Set("HX-Redirect", "/"+p.Slug)
	w.WriteHeader(201)
}

// --- API Handlers ---

func (h *Handler) apiJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
func (h *Handler) apiErr(w http.ResponseWriter, code int, msg string) {
	h.apiJSON(w, code, map[string]string{"error": msg})
}

func (h *Handler) APICreate(w http.ResponseWriter, r *http.Request) {
	var req models.PasteCreateRequest
	json.NewDecoder(r.Body).Decode(&req)
	ip := middleware.ClientIP(r)
	p, err := h.paste.Create(req, ip)
	if err != nil {
		code := 400
		if err == services.ErrTooLarge {
			code = 413
		}
		h.apiErr(w, code, err.Error())
		return
	}
	h.apiJSON(w, 201, p.ToResponse(h.cfg.Server.BaseURL))
}

func (h *Handler) APIGet(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.paste.Get(slug, r.URL.Query().Get("password"))
	if err != nil {
		code := 404
		if err == services.ErrNeedPassword {
			code = 403
		}
		h.apiErr(w, code, err.Error())
		return
	}
	h.apiJSON(w, 200, p.ToResponse(h.cfg.Server.BaseURL))
}

func (h *Handler) APIDelete(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if err := h.paste.Delete(slug); err != nil {
		h.apiErr(w, 404, "not found")
		return
	}
	h.apiJSON(w, 200, map[string]string{"status": "deleted"})
}

func (h *Handler) APIList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	pastes, _ := h.paste.List(limit, offset)
	var resp []models.PasteResponse
	for _, p := range pastes {
		resp = append(resp, p.ToResponse(h.cfg.Server.BaseURL))
	}
	if resp == nil {
		resp = []models.PasteResponse{}
	}
	h.apiJSON(w, 200, map[string]any{"pastes": resp, "limit": limit, "offset": offset})
}

func (h *Handler) APISearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		h.apiErr(w, 400, "q required")
		return
	}
	pastes, _ := h.paste.Search(q, 50)
	var resp []models.PasteResponse
	for _, p := range pastes {
		resp = append(resp, p.ToResponse(h.cfg.Server.BaseURL))
	}
	if resp == nil {
		resp = []models.PasteResponse{}
	}
	h.apiJSON(w, 200, map[string]any{"pastes": resp, "query": q})
}

func (h *Handler) APIStats(w http.ResponseWriter, r *http.Request) {
	s, _ := h.paste.Stats()
	h.apiJSON(w, 200, s)
}

func (h *Handler) APICreateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string `json:"name"`
		ExpiresIn string `json:"expires_in"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Name == "" {
		h.apiErr(w, 400, "name required")
		return
	}
	raw, err := auth.GenerateAPIToken()
	if err != nil {
		h.apiErr(w, 500, "failed to generate token")
		return
	}
	t := &models.APIToken{
		Name:        req.Name,
		TokenHash:   store.HashToken(raw),
		TokenPrefix: raw[:12] + "...",
		ExpiresAt:   models.ParseExpiresIn(req.ExpiresIn),
	}
	h.store.CreateToken(t)
	h.apiJSON(w, 201, map[string]any{"id": t.ID, "name": t.Name, "token": raw, "message": "Save this token - it won't be shown again!"})
}

func (h *Handler) APIDeleteToken(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.store.DeleteToken(id)
	h.apiJSON(w, 200, map[string]string{"status": "deleted"})
}


