package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/bennerhq/jethq/pkg/remap"
)

type Preferences struct {
	Theme                     Theme          `json:"theme"`
	PinChrome                 bool           `json:"pin_chrome"`
	HideHeaderBar             bool           `json:"hide_header_bar"`
	HideStatusBar             bool           `json:"hide_status_bar"`
	ChromeAnchor              ChromeAnchor   `json:"chrome_anchor"`
	ChromeAnchorMigrated      bool           `json:"chrome_anchor_migrated,omitempty"`
	ChromeLayout              ChromeLayout   `json:"chrome_layout"`
	HideCursor                bool           `json:"hide_cursor"`
	InvertScroll              bool           `json:"invert_scroll"`
	ShowPressedKeys           bool           `json:"show_pressed_keys"`
	ExperimentalGlobalHotkeys bool           `json:"experimental_global_hotkeys"`
	KeyboardRemaps            []remap.Rule   `json:"keyboard_remaps,omitempty"`
	AbsoluteSideButtonsViaRel bool           `json:"absolute_side_buttons_via_relative"`
	ScrollThrottle            ScrollThrottle `json:"scroll_throttle"`
	ScrollThrottleMs          int            `json:"scroll_throttle_ms,omitempty"`
	PointerMoveThrottleMs     int            `json:"pointer_move_throttle_ms,omitempty"`
}

var userHomeDir = os.UserHomeDir

//go:generate go tool github.com/dmarkham/enumer -type=Theme,ChromeAnchor,ChromeLayout,ScrollThrottle -linecomment -json -text -output prefs_enums.go

type Theme uint8

const (
	themeUnknown Theme = iota // unknown
	themeSystem               // system
	themeDark                 // dark
	themeLight                // light
)

type ChromeAnchor uint8

const (
	chromeAnchorUnknown      ChromeAnchor = iota // unknown
	chromeAnchorTopLeft                          // top_left
	chromeAnchorTopCenter                        // top_center
	chromeAnchorTopRight                         // top_right
	chromeAnchorLeftCenter                       // left_center
	chromeAnchorRightCenter                      // right_center
	chromeAnchorBottomLeft                       // bottom_left
	chromeAnchorBottomCenter                     // bottom_center
	chromeAnchorBottomRight                      // bottom_right
)

type ChromeLayout uint8

const (
	chromeLayoutUnknown    ChromeLayout = iota // unknown
	chromeLayoutHorizontal                     // horizontal
	chromeLayoutVertical                       // vertical
)

type ScrollThrottle uint8

const (
	scrollThrottleUnknown ScrollThrottle = iota // unknown
	scrollThrottleOff                           // 0
	scrollThrottle10ms                          // 10
	scrollThrottle25ms                          // 25
	scrollThrottle50ms                          // 50
	scrollThrottle100ms                         // 100
)

func defaultPreferences() Preferences {
	return Preferences{
		Theme:                     themeSystem,
		PinChrome:                 false,
		HideHeaderBar:             false,
		HideStatusBar:             false,
		ChromeAnchor:              chromeAnchorBottomRight,
		ChromeAnchorMigrated:      true,
		ChromeLayout:              chromeLayoutHorizontal,
		HideCursor:                false,
		InvertScroll:              false,
		ShowPressedKeys:           false,
		ExperimentalGlobalHotkeys: false,
		AbsoluteSideButtonsViaRel: true,
		ScrollThrottle:            scrollThrottleOff,
		ScrollThrottleMs:          0,
		PointerMoveThrottleMs:     8,
	}
}

func loadPreferences() Preferences {
	path, err := preferencesPath()
	if err != nil {
		return defaultPreferences()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		prefs := defaultPreferences()
		if errors.Is(err, os.ErrNotExist) {
			_ = savePreferences(prefs)
		}
		return prefs
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return defaultPreferences()
	}
	prefs := defaultPreferences()
	if err := json.Unmarshal(data, &prefs); err != nil {
		return defaultPreferences()
	}
	if _, ok := raw["chrome_anchor_migrated"]; !ok {
		if _, hasAnchor := raw["chrome_anchor"]; hasAnchor {
			prefs.ChromeAnchorMigrated = false
		}
	}
	if _, ok := raw["scroll_throttle_ms"]; !ok {
		if _, hasLegacyThrottle := raw["scroll_throttle"]; hasLegacyThrottle {
			prefs.ScrollThrottleMs = int(scrollThrottleFromPref(prefs.ScrollThrottle) / time.Millisecond)
		}
	}
	prefs.normalize()
	if stored, err := json.MarshalIndent(preferencesForStorage(prefs), "", "  "); err == nil && !bytes.Equal(bytes.TrimSpace(data), stored) {
		_ = savePreferences(prefs)
	}
	return prefs
}

func savePreferences(prefs Preferences) error {
	path, err := preferencesPath()
	if err != nil {
		return err
	}
	prefs.normalize()
	data, err := json.MarshalIndent(preferencesForStorage(prefs), "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func preferencesForStorage(prefs Preferences) map[string]any {
	defaults := defaultPreferences()
	stored := make(map[string]any)
	if prefs.Theme != defaults.Theme {
		stored["theme"] = prefs.Theme
	}
	if prefs.PinChrome != defaults.PinChrome {
		stored["pin_chrome"] = prefs.PinChrome
	}
	if prefs.HideHeaderBar != defaults.HideHeaderBar {
		stored["hide_header_bar"] = prefs.HideHeaderBar
	}
	if prefs.HideStatusBar != defaults.HideStatusBar {
		stored["hide_status_bar"] = prefs.HideStatusBar
	}
	if prefs.ChromeAnchor != defaults.ChromeAnchor {
		stored["chrome_anchor"] = prefs.ChromeAnchor
	}
	if prefs.ChromeAnchorMigrated != defaults.ChromeAnchorMigrated {
		stored["chrome_anchor_migrated"] = prefs.ChromeAnchorMigrated
	}
	if prefs.ChromeLayout != defaults.ChromeLayout {
		stored["chrome_layout"] = prefs.ChromeLayout
	}
	if prefs.HideCursor != defaults.HideCursor {
		stored["hide_cursor"] = prefs.HideCursor
	}
	if prefs.InvertScroll != defaults.InvertScroll {
		stored["invert_scroll"] = prefs.InvertScroll
	}
	if prefs.ShowPressedKeys != defaults.ShowPressedKeys {
		stored["show_pressed_keys"] = prefs.ShowPressedKeys
	}
	if prefs.ExperimentalGlobalHotkeys != defaults.ExperimentalGlobalHotkeys {
		stored["experimental_global_hotkeys"] = prefs.ExperimentalGlobalHotkeys
	}
	if len(prefs.KeyboardRemaps) > 0 {
		stored["keyboard_remaps"] = prefs.KeyboardRemaps
	}
	if prefs.AbsoluteSideButtonsViaRel != defaults.AbsoluteSideButtonsViaRel {
		stored["absolute_side_buttons_via_relative"] = prefs.AbsoluteSideButtonsViaRel
	}
	if prefs.ScrollThrottle != defaults.ScrollThrottle {
		stored["scroll_throttle"] = prefs.ScrollThrottle
	}
	if prefs.ScrollThrottleMs != defaults.ScrollThrottleMs {
		stored["scroll_throttle_ms"] = prefs.ScrollThrottleMs
	}
	if prefs.PointerMoveThrottleMs != defaults.PointerMoveThrottleMs {
		stored["pointer_move_throttle_ms"] = prefs.PointerMoveThrottleMs
	}
	return stored
}

func preferencesPath() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "jethq", "confg.json"), nil
}

func (p *Preferences) normalize() {
	if p.Theme == themeUnknown {
		p.Theme = themeSystem
	}
	switch p.ScrollThrottle {
	case scrollThrottleOff, scrollThrottle10ms, scrollThrottle25ms, scrollThrottle50ms, scrollThrottle100ms:
	default:
		p.ScrollThrottle = scrollThrottleOff
	}
	p.ScrollThrottleMs = clampInt(p.ScrollThrottleMs, 0, maxScrollThrottleMs)
	p.PointerMoveThrottleMs = clampInt(p.PointerMoveThrottleMs, 0, maxPointerMoveThrottleMs)
	switch p.ChromeAnchor {
	case chromeAnchorTopLeft, chromeAnchorTopCenter, chromeAnchorTopRight, chromeAnchorLeftCenter, chromeAnchorRightCenter, chromeAnchorBottomLeft, chromeAnchorBottomCenter, chromeAnchorBottomRight:
	default:
		p.ChromeAnchor = chromeAnchorBottomRight
	}
	switch p.ChromeLayout {
	case chromeLayoutHorizontal, chromeLayoutVertical:
	default:
		p.ChromeLayout = chromeLayoutHorizontal
	}
}

func scrollThrottleFromPref(value ScrollThrottle) time.Duration {
	switch value {
	case scrollThrottle10ms:
		return 10 * time.Millisecond
	case scrollThrottle25ms:
		return 25 * time.Millisecond
	case scrollThrottle50ms:
		return 50 * time.Millisecond
	case scrollThrottle100ms:
		return 100 * time.Millisecond
	default:
		return 0
	}
}

func scrollThrottlePref(value time.Duration) ScrollThrottle {
	switch value {
	case 10 * time.Millisecond:
		return scrollThrottle10ms
	case 25 * time.Millisecond:
		return scrollThrottle25ms
	case 50 * time.Millisecond:
		return scrollThrottle50ms
	case 100 * time.Millisecond:
		return scrollThrottle100ms
	default:
		return scrollThrottleOff
	}
}

func throttleDurationFromMs(value int) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func clampInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
