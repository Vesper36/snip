package i18n

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed en.json zh.json
var fs embed.FS

var (
	translations = map[string]map[string]string{}
	DefaultLang  = "en"
	supported    = []string{"en", "zh"}
	Labels       = map[string]string{
		"en": "English",
		"zh": "中文",
	}
)

func init() {
	for _, lang := range supported {
		data, err := fs.ReadFile(lang + ".json")
		if err != nil {
			continue
		}
		m := map[string]string{}
		json.Unmarshal(data, &m)
		translations[lang] = m
	}
}

// T returns the translation for the given key in the specified language.
func T(lang, key string) string {
	if m, ok := translations[lang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, ok := translations[DefaultLang]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}

// Supported returns the list of supported language codes.
func Supported() []string {
	return supported
}

// DetectLanguage extracts language from Accept-Language header.
func DetectLanguage(acceptLang string) string {
	if acceptLang == "" {
		return DefaultLang
	}
	for _, part := range strings.Split(acceptLang, ",") {
		lang := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		code := strings.ToLower(strings.SplitN(lang, "-", 2)[0])
		for _, s := range supported {
			if code == s {
				return code
			}
		}
	}
	return DefaultLang
}
