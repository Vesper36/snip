package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vesper/snip/internal/models"
)

// Aliases for test readability
type Paste = models.Paste
type APIToken = models.APIToken

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir, err := os.MkdirTemp("", "snip-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	s, err := New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStoreCRUD(t *testing.T) {
	s := newTestStore(t)

	p := &Paste{
		Slug:     "test1",
		Title:    "Test Paste",
		Content:  "hello world",
		Language: "go",
		FileSize: 11,
	}
	if err := s.CreatePaste(p); err != nil {
		t.Fatal(err)
	}
	if p.ID == 0 {
		t.Error("ID should be set after create")
	}

	got, err := s.GetPaste("test1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("paste not found")
	}
	if got.Title != "Test Paste" {
		t.Errorf("Title = %q", got.Title)
	}

	if err := s.IncrementViews("test1"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.GetPaste("test1")
	if got.Views != 1 {
		t.Errorf("Views = %d, want 1", got.Views)
	}

	if err := s.DeletePaste("test1"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.GetPaste("test1")
	if got != nil {
		t.Error("paste should be deleted")
	}
}

func TestStoreList(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 5; i++ {
		s.CreatePaste(&Paste{
			Slug:     "p" + string(rune('a'+i)),
			Content:  "test",
			Language: "plaintext",
		})
	}

	pastes, err := s.ListPastes(10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(pastes) != 5 {
		t.Errorf("got %d pastes, want 5", len(pastes))
	}
}

func TestStoreCount(t *testing.T) {
	s := newTestStore(t)

	n, _ := s.CountPastes()
	if n != 0 {
		t.Errorf("initial count = %d, want 0", n)
	}

	s.CreatePaste(&Paste{Slug: "x", Content: "hi"})
	n, _ = s.CountPastes()
	if n != 1 {
		t.Errorf("count = %d, want 1", n)
	}
}

func TestStoreExpiredCleanup(t *testing.T) {
	s := newTestStore(t)

	past := time.Now().Add(-time.Hour)
	s.CreatePaste(&Paste{
		Slug:      "old",
		Content:   "old content",
		ExpiresAt: &past,
	})
	s.CreatePaste(&Paste{Slug: "fresh", Content: "fresh"})

	deleted, err := s.CleanupExpired()
	if err != nil {
		t.Fatal(err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}

	if p, _ := s.GetPaste("fresh"); p == nil {
		t.Error("fresh paste should not be deleted")
	}
}

func TestStoreAPIToken(t *testing.T) {
	s := newTestStore(t)

	tk := &APIToken{
		Name:        "test-token",
		TokenHash:   HashToken("abc123"),
		TokenPrefix: "abc123...",
	}
	if err := s.CreateToken(tk); err != nil {
		t.Fatal(err)
	}

	got, err := s.GetTokenByHash(HashToken("abc123"))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Name != "test-token" {
		t.Error("token not found")
	}

	// Use the token
	if err := s.UpdateTokenUsed(tk.ID); err != nil {
		t.Fatal(err)
	}
	got, _ = s.GetTokenByHash(HashToken("abc123"))
	if got.LastUsedAt == nil {
		t.Error("last_used_at should be set")
	}
}

func TestHashToken(t *testing.T) {
	h1 := HashToken("token1")
	h2 := HashToken("token2")
	if h1 == h2 {
		t.Error("different tokens should have different hashes")
	}
	if h1 != HashToken("token1") {
		t.Error("same token should have same hash (deterministic)")
	}
	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64 (sha256 hex)", len(h1))
	}
}

func TestStoreStats(t *testing.T) {
	s := newTestStore(t)

	s.CreatePaste(&Paste{Slug: "p1", Content: "hi", FileSize: 2})
	s.CreatePaste(&Paste{Slug: "p2", Content: "longer", FileSize: 6})
	s.CreateToken(&APIToken{Name: "t", TokenHash: "h", TokenPrefix: "h"})

	stats, err := s.GetStats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalPastes != 2 {
		t.Errorf("TotalPastes = %d, want 2", stats.TotalPastes)
	}
	if stats.TotalBytes != 8 {
		t.Errorf("TotalBytes = %d, want 8", stats.TotalBytes)
	}
	if stats.TotalTokens != 1 {
		t.Errorf("TotalTokens = %d, want 1", stats.TotalTokens)
	}
}
