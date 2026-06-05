package models

import (
	"testing"
	"time"
)

func TestGenerateSlug_Length(t *testing.T) {
	slug := GenerateSlug(8)
	if len(slug) != 8 {
		t.Fatalf("Expected slug length 8, got %d", len(slug))
	}
}

func TestGenerateSlug_Uniqueness(t *testing.T) {
	slugs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		slug := GenerateSlug(8)
		if slugs[slug] {
			t.Fatalf("Duplicate slug generated: %s", slug)
		}
		slugs[slug] = true
	}
}

func TestPaste_IsExpired(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		paste    *Paste
		expected bool
	}{
		{"no expiry", &Paste{}, false},
		{"expired", &Paste{ExpiresAt: &past}, true},
		{"not expired", &Paste{ExpiresAt: &future}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.paste.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPaste_IsViewLimitReached(t *testing.T) {
	tests := []struct {
		name     string
		paste    *Paste
		expected bool
	}{
		{"no limit", &Paste{MaxViews: 0, Views: 100}, false},
		{"under limit", &Paste{MaxViews: 10, Views: 5}, false},
		{"at limit", &Paste{MaxViews: 10, Views: 10}, true},
		{"over limit", &Paste{MaxViews: 10, Views: 15}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.paste.IsViewLimitReached(); got != tt.expected {
				t.Errorf("IsViewLimitReached() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPaste_ToResponse(t *testing.T) {
	p := &Paste{
		Slug:     "abc123",
		Title:    "Test",
		Content:  "hello",
		Language: "go",
		FileSize: 5,
		Views:    3,
	}
	resp := p.ToResponse("https://example.com")
	if resp.Slug != "abc123" {
		t.Errorf("Slug = %s, want abc123", resp.Slug)
	}
	if resp.URL != "https://example.com/abc123" {
		t.Errorf("URL = %s, want https://example.com/abc123", resp.URL)
	}
	if resp.RawURL != "https://example.com/abc123/raw" {
		t.Errorf("RawURL = %s, want https://example.com/abc123/raw", resp.RawURL)
	}
}

func TestParseExpiresIn(t *testing.T) {
	tests := []struct {
		input    string
		isNil    bool
		minDur   time.Duration
		maxDur   time.Duration
	}{
		{"never", true, 0, 0},
		{"10m", false, 9 * time.Minute, 11 * time.Minute},
		{"1h", false, 59 * time.Minute, 61 * time.Minute},
		{"1d", false, 23 * time.Hour, 25 * time.Hour},
		{"1w", false, 6 * 24 * time.Hour, 8 * 24 * time.Hour},
		{"1m", false, 29 * 24 * time.Hour, 31 * 24 * time.Hour},
		{"invalid", true, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseExpiresIn(tt.input)
			if tt.isNil {
				if result != nil {
					t.Errorf("ParseExpiresIn(%q) = %v, want nil", tt.input, result)
				}
				return
			}
			if result == nil {
				t.Fatalf("ParseExpiresIn(%q) = nil, want non-nil", tt.input)
			}
			d := time.Until(*result)
			if d < tt.minDur || d > tt.maxDur {
				t.Errorf("ParseExpiresIn(%q) duration = %v, want between %v and %v", tt.input, d, tt.minDur, tt.maxDur)
			}
		})
	}
}

func TestLanguageList(t *testing.T) {
	langs := LanguageList()
	if len(langs) == 0 {
		t.Fatal("LanguageList is empty")
	}
	// Check some expected languages
	found := map[string]bool{}
	for _, l := range langs {
		found[l] = true
	}
	for _, expected := range []string{"go", "python", "javascript", "typescript", "rust", "java"} {
		if !found[expected] {
			t.Errorf("Missing expected language: %s", expected)
		}
	}
}
