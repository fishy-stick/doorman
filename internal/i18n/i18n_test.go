package i18n

import (
	"net/http/httptest"
	"testing"
)

func TestLocaleFromRequestPrefersDoormanHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/admin/api/session", nil)
	req.Header.Set(HeaderDoormanLocale, "zh-CN")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	if locale := LocaleFromRequest(req); locale != LocaleSimplifiedChinese {
		t.Fatalf("LocaleFromRequest() = %q, want %q", locale, LocaleSimplifiedChinese)
	}
}

func TestLocaleFromRequestFallsBackToAcceptLanguage(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/admin/api/session", nil)
	req.Header.Set("Accept-Language", "fr-FR, zh-CN;q=0.9, en;q=0.8")

	if locale := LocaleFromRequest(req); locale != LocaleSimplifiedChinese {
		t.Fatalf("LocaleFromRequest() = %q, want %q", locale, LocaleSimplifiedChinese)
	}
}

func TestLocaleFromRequestDefaultsToEnglish(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "/admin/api/session", nil)
	req.Header.Set("Accept-Language", "fr-FR, de-DE;q=0.9")

	if locale := LocaleFromRequest(req); locale != LocaleEnglish {
		t.Fatalf("LocaleFromRequest() = %q, want %q", locale, LocaleEnglish)
	}
}
