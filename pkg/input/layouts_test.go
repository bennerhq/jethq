package input

import "testing"

func TestSupportedKeyboardLayoutsMatchesUpstreamCatalog(t *testing.T) {
	want := []KeyboardLayout{
		{Code: "cs-CZ", Label: "Čeština"},
		{Code: "da-DK", Label: "Dansk"},
		{Code: "de-CH", Label: "Schwiizerdütsch"},
		{Code: "de-DE", Label: "Deutsch"},
		{Code: "en-UK", Label: "English (UK)"},
		{Code: "en-US", Label: "English (US)"},
		{Code: "es-ES", Label: "Español"},
		{Code: "nl-BE", Label: "Belgisch Nederlands"},
		{Code: "fr-CH", Label: "Français de Suisse"},
		{Code: "fr-FR", Label: "Français"},
		{Code: "hu-HU", Label: "Magyar"},
		{Code: "it-IT", Label: "Italiano"},
		{Code: "ja-JP", Label: "Japanese"},
		{Code: "nb-NO", Label: "Norsk bokmål"},
		{Code: "pl-PL", Label: "Polski"},
		{Code: "pt-PT", Label: "Português"},
		{Code: "sv-SE", Label: "Svenska"},
		{Code: "sl-SI", Label: "Slovenian"},
		{Code: "ru-RU", Label: "Русская"},
	}

	got := SupportedKeyboardLayouts()
	if len(got) != len(want) {
		t.Fatalf("layout count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("layout %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestNormalizeKeyboardLayoutCodeAcceptsUpstreamCodes(t *testing.T) {
	tests := map[string]string{
		"":        "",
		"sv-SE":   "sv-SE",
		"sv_SE":   "sv-SE",
		" nb-NO ": "nb-NO",
		"nl-BE":   "nl-BE",
		"bad":     "en-US",
	}

	for in, want := range tests {
		if got := NormalizeKeyboardLayoutCode(in); got != want {
			t.Fatalf("NormalizeKeyboardLayoutCode(%q) = %q, want %q", in, got, want)
		}
	}
}
