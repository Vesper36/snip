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
	"github.com/vesper/snip/internal/i18n"
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
		"timeAgo": func(t time.Time, lang ...string) string {
			d := time.Since(t)
			l := "en"
			if len(lang) > 0 { l = lang[0] }
			switch {
			case d < time.Minute:
				return i18n.T(l, "just_now")
			case d < time.Hour:
				return fmt.Sprintf("%d%s", int(d.Minutes()), i18n.T(l, "min_ago"))
			case d < 24*time.Hour:
				return fmt.Sprintf("%d%s", int(d.Hours()), i18n.T(l, "hr_ago"))
			default:
				return fmt.Sprintf("%d%s", int(d.Hours()/24), i18n.T(l, "day_ago"))
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

func (h *Handler) render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Inject i18n data
	lang := i18n.DefaultLang
	if v := r.Context().Value(middleware.CtxLang); v != nil {
		lang = v.(string)
	}
	data["Lang"] = lang
	data["T"] = func(key string) string { return i18n.T(lang, key) }
	data["SupportedLangs"] = i18n.Supported()
	data["LangLabels"] = i18n.Labels

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

func (h *Handler) errPage(w http.ResponseWriter, r *http.Request, code int, title, msg string) {
	w.WriteHeader(code)
	h.render(w, r, "error.html", map[string]any{"Code": code, "Title": title, "Message": msg})
}

// --- Web Handlers ---

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "home.html", map[string]any{
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
			h.errPage(w, r, 404, "error_not_found_title", "error_not_found_msg")
		case services.ErrExpired:
			h.errPage(w, r, 410, "error_expired_title", "error_expired_msg")
		case services.ErrMaxViews:
			h.errPage(w, r, 410, "error_gone_title", "error_gone_msg")
		case services.ErrNeedPassword:
			h.render(w, r, "password.html", map[string]any{"Slug": slug, "NeedPW": true})
		case services.ErrBadPassword:
			h.render(w, r, "password.html", map[string]any{"Slug": slug, "NeedPW": true, "Error": "auth_wrong_password"})
		default:
			h.errPage(w, r, 500, "error_generic_title", "error_generic_msg")
		}
		return
	}
	h.render(w, r, "view.html", map[string]any{"P": p, "BaseURL": h.cfg.Server.BaseURL, "Languages": models.LanguageList()})
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

// Fork creates a new paste based on an existing one.
func (h *Handler) Fork(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	ip := middleware.ClientIP(r)
	p, err := h.paste.Fork(slug, ip)
	if err != nil {
		if err == services.ErrNotFound {
			h.errPage(w, r, 404, "error_not_found_title", "error_not_found_msg")
		} else {
			h.errPage(w, r, 500, "error_generic_title", "error_generic_msg")
		}
		return
	}
	http.Redirect(w, r, "/"+p.Slug, http.StatusSeeOther)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Redirect(w, r, "/", 302)
		return
	}
	pastes, _ := h.paste.Search(q, 50)
	h.render(w, r, "search.html", map[string]any{"Query": q, "Pastes": pastes, "BaseURL": h.cfg.Server.BaseURL})
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
	h.render(w, r, "list.html", map[string]any{
		"Pastes": pastes, "Page": page, "HasNext": hasNext, "Total": total, "BaseURL": h.cfg.Server.BaseURL,
	})
}

func (h *Handler) Settings(w http.ResponseWriter, r *http.Request) {
	tokens, _ := h.store.ListTokens()
	stats, _ := h.paste.Stats()
	h.render(w, r, "settings.html", map[string]any{"Tokens": tokens, "Stats": stats, "BaseURL": h.cfg.Server.BaseURL})
}

// --- HTMX Handler ---

// SetLang handles language switching via cookie.
func (h *Handler) SetLang(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = r.FormValue("lang")
	}
	// Validate
	valid := false
	for _, s := range i18n.Supported() {
		if lang == s {
			valid = true
			break
		}
	}
	if !valid {
		lang = i18n.DefaultLang
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	// Redirect back
	referer := r.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	http.Redirect(w, r, referer, http.StatusSeeOther)
}

func (h *Handler) CreateHTMX(w http.ResponseWriter, r *http.Request) {
	req := models.PasteCreateRequest{
		Title:      r.FormValue("title"),
		Content:    r.FormValue("content"),
		Language:   r.FormValue("language"),
		Password:   r.FormValue("password"),
		ExpiresIn:  r.FormValue("expires_in"),
		CustomSlug: r.FormValue("custom_slug"),
	}
	if r.FormValue("burn_after_read") == "on" {
		req.BurnAfterRead = true
	}
	if v := r.FormValue("max_views"); v != "" {
		req.MaxViews, _ = strconv.Atoi(v)
	}

	// Handle file upload
	if file, header, err := r.FormFile("file"); err == nil {
		defer file.Close()
		buf := make([]byte, h.cfg.Paste.MaxSize)
		n, _ := file.Read(buf)
		req.Content = string(buf[:n])
		if req.Title == "" {
			req.Title = header.Filename
		}
		if req.Language == "" || req.Language == "auto" {
			req.Language = "plaintext"
		}
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
	// Accept both JSON and form data (HTMX)
	ct := r.Header.Get("Content-Type")
	if len(ct) >= 5 && ct[:5] == "multi" {
		_ = r.ParseForm()
		req.Name = r.FormValue("name")
		req.ExpiresIn = r.FormValue("expires_in")
	} else if len(ct) >= 33 && ct[:33] == "application/x-www-form-urlencoded" {
		_ = r.ParseForm()
		req.Name = r.FormValue("name")
		req.ExpiresIn = r.FormValue("expires_in")
	} else {
		json.NewDecoder(r.Body).Decode(&req)
	}
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

	// If HTMX request, return HTML; else JSON
	if r.Header.Get("HX-Request") == "true" {
		lang := i18n.DefaultLang
		if v := r.Context().Value(middleware.CtxLang); v != nil {
			lang = v.(string)
		}
		html := `<div class="alert alert-ok" style="background:rgba(34,197,94,0.08);color:var(--green);border:1px solid rgba(34,197,94,0.2);padding:0.875rem 1.125rem;border-radius:8px;font-size:0.875rem;margin-bottom:1rem;display:flex;align-items:center;gap:0.5rem">
			<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg>
			` + i18n.T(lang, "token_save_hint") + ` Token: <code style="background:var(--bg);padding:0.125rem 0.5rem;border-radius:4px;color:var(--accent);font-family:monospace">` + raw + `</code>
		</div>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(201)
		w.Write([]byte(html))
		return
	}
	h.apiJSON(w, 201, map[string]any{"id": t.ID, "name": t.Name, "token": raw, "message": "Save this token - it won't be shown again!"})
}

func (h *Handler) APIDeleteToken(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	h.store.DeleteToken(id)
	h.apiJSON(w, 200, map[string]string{"status": "deleted"})
}

// Healthz is a health check endpoint (no auth required).
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := h.store.Ping(r.Context()); err != nil {
		w.WriteHeader(503)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}
	stats, _ := h.paste.Stats()
	json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"time":     time.Now().UTC().Format(time.RFC3339),
		"version":  "1.0.0",
		"pastes":   stats.TotalPastes,
		"views":    stats.TotalViews,
		"tokens":   stats.TotalTokens,
		"storage":  stats.TotalBytes,
	})
}

// Metrics exposes Prometheus-compatible metrics.
func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	stats, _ := h.paste.Stats()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP snip_pastes_total Total number of pastes\n")
	fmt.Fprintf(w, "# TYPE snip_pastes_total gauge\n")
	fmt.Fprintf(w, "snip_pastes_total %d\n", stats.TotalPastes)
	fmt.Fprintf(w, "# HELP snip_views_total Total view count\n")
	fmt.Fprintf(w, "# TYPE snip_views_total counter\n")
	fmt.Fprintf(w, "snip_views_total %d\n", stats.TotalViews)
	fmt.Fprintf(w, "# HELP snip_tokens_total Total API tokens\n")
	fmt.Fprintf(w, "# TYPE snip_tokens_total gauge\n")
	fmt.Fprintf(w, "snip_tokens_total %d\n", stats.TotalTokens)
	fmt.Fprintf(w, "# HELP snip_storage_bytes Total storage used\n")
	fmt.Fprintf(w, "# TYPE snip_storage_bytes gauge\n")
	fmt.Fprintf(w, "snip_storage_bytes %d\n", stats.TotalBytes)
}


