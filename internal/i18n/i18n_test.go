package i18n

import "testing"

func TestT_ExistingKey(t *testing.T) {
	v := T("en", "brand")
	if v != "Snip" {
		t.Fatalf("Expected 'Snip', got '%s'", v)
	}
}

func TestT_ZHTranslation(t *testing.T) {
	v := T("zh", "brand")
	if v != "Snip" {
		t.Fatalf("Expected 'Snip', got '%s'", v)
	}
}

func TestT_JATranslation(t *testing.T) {
	v := T("ja", "brand")
	if v != "Snip" {
		t.Fatalf("Expected 'Snip', got '%s'", v)
	}
}

func TestT_FRTranslation(t *testing.T) {
	v := T("fr", "brand")
	if v != "Snip" {
		t.Fatalf("Expected 'Snip', got '%s'", v)
	}
}

func TestT_DETranslation(t *testing.T) {
	v := T("de", "brand")
	if v != "Snip" {
		t.Fatalf("Expected 'Snip', got '%s'", v)
	}
}

func TestT_MissingKey_FallbackToEN(t *testing.T) {
	v := T("zh", "nonexistent_key_xyz")
	if v != "nonexistent_key_xyz" {
		t.Fatalf("Expected key itself as fallback, got '%s'", v)
	}
}

func TestT_UnknownLang_FallbackToEN(t *testing.T) {
	v := T("xx", "brand")
	if v != "Snip" {
		t.Fatalf("Expected 'Snip' from EN fallback, got '%s'", v)
	}
}

func TestSupported(t *testing.T) {
	supported := Supported()
	if len(supported) != 5 {
		t.Fatalf("Expected 5 supported languages, got %d", len(supported))
	}
	found := map[string]bool{}
	for _, s := range supported {
		found[s] = true
	}
	for _, expected := range []string{"en", "zh", "ja", "fr", "de"} {
		if !found[expected] {
			t.Fatalf("Missing supported language: %s", expected)
		}
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "en"},
		{"en-US,en;q=0.9", "en"},
		{"zh-CN,zh;q=0.9", "zh"},
		{"ja-JP,ja;q=0.9", "ja"},
		{"fr-FR,fr;q=0.9", "fr"},
		{"de-DE,de;q=0.9", "de"},
		{"ko-KR,ko;q=0.9", "en"}, // unsupported falls back to en
	}
	for _, tt := range tests {
		got := DetectLanguage(tt.input)
		if got != tt.expected {
			t.Errorf("DetectLanguage(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
