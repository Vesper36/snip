package services

import (
	"testing"

	"github.com/vesper/snip/internal/auth"
	"github.com/vesper/snip/internal/models"
)

func TestSlugGeneration(t *testing.T) {
	s := models.GenerateSlug(8)
	if len(s) != 8 {
		t.Errorf("expected slug length 8, got %d (%s)", len(s), s)
	}
}

func TestSlugUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		s := models.GenerateSlug(8)
		if seen[s] {
			t.Fatalf("duplicate slug generated: %s", s)
		}
		seen[s] = true
	}
}

func TestParseExpiresIn(t *testing.T) {
	tests := []struct {
		input   string
		wantNil bool
	}{
		{"10m", false},
		{"1h", false},
		{"1d", false},
		{"1w", false},
		{"1m", false},
		{"never", true},
		{"", true},
		{"invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := models.ParseExpiresIn(tt.input)
			if (got == nil) != tt.wantNil {
				t.Errorf("ParseExpiresIn(%q) nil=%v, want %v", tt.input, got == nil, tt.wantNil)
			}
		})
	}
}

func TestPasswordHashing(t *testing.T) {
	pw := "test-password-123"
	hash, err := auth.HashPassword(pw)
	if err != nil {
		t.Fatal(err)
	}
	if !auth.CheckPassword(pw, hash) {
		t.Error("password should match hash")
	}
	if auth.CheckPassword("wrong-password", hash) {
		t.Error("wrong password should not match")
	}
}

func TestAPITokenGeneration(t *testing.T) {
	t1, err := auth.GenerateAPIToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(t1) < 20 {
		t.Errorf("token too short: %s", t1)
	}
	// Each should be unique
	t2, _ := auth.GenerateAPIToken()
	if t1 == t2 {
		t.Error("consecutive tokens should differ")
	}
}

func TestLanguageDetection(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`{"name": "value"}`, "json"},
		{"package main\nfunc main() {}", "go"},
		{"def hello():\n    print('hi')", "python"},
		{"<html><body></body></html>", "html"},
		{"# Heading\nText", "markdown"},
		{"#!/bin/bash\necho hi", "bash"},
	}
	for _, tt := range tests {
		got := detectLanguage(tt.input)
		if got != tt.expected {
			t.Errorf("detectLanguage(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestPasteIsExpired(t *testing.T) {
	p := &models.Paste{}
	if p.IsExpired() {
		t.Error("nil expiry should not be expired")
	}
}

func TestPasteIsViewLimitReached(t *testing.T) {
	p := &models.Paste{MaxViews: 0, Views: 100}
	if p.IsViewLimitReached() {
		t.Error("max_views=0 should be unlimited")
	}
	p.MaxViews = 5
	p.Views = 5
	if !p.IsViewLimitReached() {
		t.Error("views == max_views should be reached")
	}
	p.Views = 4
	if p.IsViewLimitReached() {
		t.Error("views < max_views should not be reached")
	}
}

func TestPasteResponse(t *testing.T) {
	p := &models.Paste{
		Slug:   "abc123",
		Title:  "Test",
		Views:  42,
		Content: "hello",
	}
	resp := p.ToResponse("https://example.com")
	if resp.URL != "https://example.com/abc123" {
		t.Errorf("URL = %q", resp.URL)
	}
	if resp.RawURL != "https://example.com/abc123/raw" {
		t.Errorf("RawURL = %q", resp.RawURL)
	}
}
