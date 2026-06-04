package services

import (
	"errors"
	"strings"

	"github.com/vesper/snip/internal/auth"
	"github.com/vesper/snip/internal/models"
	"github.com/vesper/snip/internal/store"
)

var (
	ErrNotFound      = errors.New("paste not found")
	ErrExpired       = errors.New("paste expired")
	ErrMaxViews      = errors.New("max views reached")
	ErrNeedPassword  = errors.New("password required")
	ErrBadPassword   = errors.New("wrong password")
	ErrSlugTaken     = errors.New("slug already taken")
	ErrTooLarge      = errors.New("content too large")
	ErrEmptyContent  = errors.New("content is required")
)

type PasteService struct {
	store   *store.Store
	maxSize int64
	baseURL string
}

func NewPasteService(s *store.Store, maxSize int64, baseURL string) *PasteService {
	return &PasteService{store: s, maxSize: maxSize, baseURL: baseURL}
}

func (s *PasteService) Create(req models.PasteCreateRequest, ip string) (*models.Paste, error) {
	if strings.TrimSpace(req.Content) == "" {
		return nil, ErrEmptyContent
	}
	if int64(len(req.Content)) > s.maxSize {
		return nil, ErrTooLarge
	}

	slug := req.CustomSlug
	if slug == "" {
		slug = models.GenerateSlug(8)
	} else {
		if existing, _ := s.store.GetPaste(slug); existing != nil {
			return nil, ErrSlugTaken
		}
	}

	lang := req.Language
	if lang == "" || lang == "auto" {
		lang = detectLanguage(req.Content)
	}

	var pwHash string
	if req.Password != "" {
		h, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		pwHash = h
	}

	p := &models.Paste{
		Slug:          slug,
		Title:         req.Title,
		Content:       req.Content,
		Language:      lang,
		FileSize:      int64(len(req.Content)),
		IsEncrypted:   false,
		PasswordHash:  pwHash,
		BurnAfterRead: req.BurnAfterRead,
		ExpiresAt:     models.ParseExpiresIn(req.ExpiresIn),
		MaxViews:      req.MaxViews,
		AuthorIP:      ip,
	}
	if err := s.store.CreatePaste(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PasteService) Get(slug, password string) (*models.Paste, error) {
	p, err := s.store.GetPaste(slug)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrNotFound
	}
	if p.IsExpired() {
		s.store.DeletePaste(slug)
		return nil, ErrExpired
	}
	if p.IsViewLimitReached() {
		return nil, ErrMaxViews
	}
	if p.PasswordHash != "" {
		if password == "" {
			return nil, ErrNeedPassword
		}
		if !auth.CheckPassword(password, p.PasswordHash) {
			return nil, ErrBadPassword
		}
	}
	s.store.IncrementViews(slug)
	if p.BurnAfterRead {
		go s.store.DeletePaste(slug)
	}
	return p, nil
}

func (s *PasteService) Delete(slug string) error {
	return s.store.DeletePaste(slug)
}

func (s *PasteService) List(limit, offset int) ([]*models.Paste, error) {
	return s.store.ListPastes(limit, offset)
}

func (s *PasteService) Search(q string, limit int) ([]*models.Paste, error) {
	return s.store.SearchPastes(q, limit)
}

func (s *PasteService) Stats() (*store.Stats, error) {
	return s.store.GetStats()
}

func (s *PasteService) Cleanup() (int64, error) {
	return s.store.CleanupExpired()
}

func detectLanguage(c string) string {
	t := strings.TrimSpace(c)
	if (strings.HasPrefix(t, "{") || strings.HasPrefix(t, "[")) && strings.Contains(t, "\"") {
		return "json"
	}
	if strings.HasPrefix(t, "<!DOCTYPE") || strings.HasPrefix(t, "<html") {
		return "html"
	}
	if strings.HasPrefix(t, "<?xml") {
		return "xml"
	}
	if strings.HasPrefix(t, "---") || (strings.Contains(t, ": ") && !strings.Contains(t, "{")) {
		return "yaml"
	}
	if strings.HasPrefix(t, "#!/") {
		if strings.Contains(t, "python") {
			return "python"
		}
		return "bash"
	}
	if strings.Contains(t, "package ") && strings.Contains(t, "func ") {
		return "go"
	}
	if strings.Contains(t, "def ") && strings.Contains(t, ":") && !strings.Contains(t, "{") {
		return "python"
	}
	if strings.Contains(t, "func ") && strings.Contains(t, "=>") {
		return "typescript"
	}
	if strings.Contains(t, "const ") || strings.Contains(t, "let ") || strings.Contains(t, "function ") {
		return "javascript"
	}
	u := strings.ToUpper(t)
	if strings.Contains(u, "SELECT ") || strings.Contains(u, "CREATE TABLE") {
		return "sql"
	}
	if strings.HasPrefix(u, "FROM ") && strings.Contains(u, "RUN ") {
		return "dockerfile"
	}
	if strings.Contains(t, "# ") || strings.Contains(t, "```") {
		return "markdown"
	}
	return "plaintext"
}
