package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Paste struct {
	ID            int64      `json:"id"`
	Slug          string     `json:"slug"`
	Title         string     `json:"title"`
	Content       string     `json:"content,omitempty"`
	Language      string     `json:"language"`
	FileSize      int64      `json:"file_size"`
	IsEncrypted   bool       `json:"is_encrypted"`
	PasswordHash  string     `json:"-"`
	BurnAfterRead bool       `json:"burn_after_read"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	MaxViews      int        `json:"max_views"`
	Views         int        `json:"views"`
	AuthorIP      string     `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type PasteCreateRequest struct {
	Title         string `json:"title" form:"title"`
	Content       string `json:"content" form:"content"`
	Language      string `json:"language" form:"language"`
	Password      string `json:"password" form:"password"`
	BurnAfterRead bool   `json:"burn_after_read" form:"burn_after_read"`
	ExpiresIn     string `json:"expires_in" form:"expires_in"`
	MaxViews      int    `json:"max_views" form:"max_views"`
	CustomSlug    string `json:"custom_slug" form:"custom_slug"`
}

type PasteResponse struct {
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Content     string     `json:"content,omitempty"`
	Language    string     `json:"language"`
	FileSize    int64      `json:"file_size"`
	IsEncrypted bool       `json:"is_encrypted"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Views       int        `json:"views"`
	CreatedAt   time.Time  `json:"created_at"`
	URL         string     `json:"url"`
	RawURL      string     `json:"raw_url"`
}

type APIToken struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	TokenHash   string     `json:"-"`
	TokenPrefix string     `json:"token_prefix"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func GenerateSlug(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

func (p *Paste) ToResponse(baseURL string) PasteResponse {
	return PasteResponse{
		Slug:        p.Slug,
		Title:       p.Title,
		Content:     p.Content,
		Language:    p.Language,
		FileSize:    p.FileSize,
		IsEncrypted: p.IsEncrypted,
		ExpiresAt:   p.ExpiresAt,
		Views:       p.Views,
		CreatedAt:   p.CreatedAt,
		URL:         baseURL + "/" + p.Slug,
		RawURL:      baseURL + "/" + p.Slug + "/raw",
	}
}

func (p *Paste) IsExpired() bool {
	return p.ExpiresAt != nil && time.Now().After(*p.ExpiresAt)
}

func (p *Paste) IsViewLimitReached() bool {
	return p.MaxViews > 0 && p.Views >= p.MaxViews
}

func ParseExpiresIn(s string) *time.Time {
	var d time.Duration
	switch s {
	case "10m":
		d = 10 * time.Minute
	case "1h":
		d = time.Hour
	case "1d":
		d = 24 * time.Hour
	case "1w":
		d = 7 * 24 * time.Hour
	case "1m":
		d = 30 * 24 * time.Hour
	default:
		return nil
	}
	t := time.Now().Add(d)
	return &t
}

func LanguageList() []string {
	return []string{
		"auto", "plaintext", "bash", "c", "cpp", "csharp", "css", "dart",
		"dockerfile", "elixir", "erlang", "go", "graphql", "haskell", "html",
		"java", "javascript", "json", "julia", "kotlin", "latex", "lua",
		"makefile", "markdown", "matlab", "nginx", "objective-c", "ocaml",
		"perl", "php", "powershell", "python", "r", "ruby", "rust", "scala",
		"scss", "shell", "sql", "swift", "terraform", "toml", "typescript",
		"vim", "xml", "yaml", "zig",
	}
}
