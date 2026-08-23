package input

import "strings"

type KeyboardLayout struct {
	Code  string
	Label string
}

var supportedKeyboardLayouts = []KeyboardLayout{
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

var supportedKeyboardLayoutCodes = func() map[string]struct{} {
	codes := make(map[string]struct{}, len(supportedKeyboardLayouts))
	for _, layout := range supportedKeyboardLayouts {
		codes[layout.Code] = struct{}{}
	}
	return codes
}()

func SupportedKeyboardLayouts() []KeyboardLayout {
	out := make([]KeyboardLayout, len(supportedKeyboardLayouts))
	copy(out, supportedKeyboardLayouts)
	return out
}

func NormalizeKeyboardLayoutCode(layout string) string {
	layout = strings.TrimSpace(layout)
	layout = strings.ReplaceAll(layout, "_", "-")
	if layout == "" {
		return layout
	}
	if _, ok := supportedKeyboardLayoutCodes[layout]; ok {
		return layout
	}
	return "en-US"
}
