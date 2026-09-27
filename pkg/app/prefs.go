package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
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
	KnownDevices              []KnownDevice  `json:"known_devices,omitempty"`
}

// KnownDevice is a JetKVM the scanner has seen before. They are replayed into
// the launcher at startup so the list is useful before the first subnet sweep
// completes, and stay marked as unconfirmed until this session's scan answers.
type KnownDevice struct {
	Name     string    `json:"name"`
	BaseURL  string    `json:"base_url"`
	Host     string    `json:"host,omitempty"`
	IP       string    `json:"ip,omitempty"`
	Scheme   string    `json:"scheme,omitempty"`
	IsSetup  bool      `json:"is_setup,omitempty"`
	LastSeen time.Time `json:"last_seen,omitempty"`
	LastUsed time.Time `json:"last_used,omitempty"`
}

// recency ranks a device for trimming: a device you connect to regularly should
// outlive one the scanner merely saw more recently.
func (d KnownDevice) recency() time.Time {
	if d.LastUsed.After(d.LastSeen) {
		return d.LastUsed
	}
	return d.LastSeen
}

const maxKnownDevices = 16

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
	migrateLegacyPreferences(path)
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
	if len(prefs.KnownDevices) > 0 {
		stored["known_devices"] = prefs.KnownDevices
	}
	return stored
}

func preferencesPath() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "jethq", "config.json"), nil
}

// legacyPreferencesFile is the misspelled name earlier releases wrote to.
const legacyPreferencesFile = "confg.json"

// migrateLegacyPreferences renames a preferences file written under the old
// misspelled name to path. An existing file at path always wins, and the
// legacy file is then left untouched.
func migrateLegacyPreferences(path string) {
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		return
	}
	legacy := filepath.Join(filepath.Dir(path), legacyPreferencesFile)
	if _, err := os.Stat(legacy); err != nil {
		return
	}
	_ = os.Rename(legacy, path)
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
	p.KnownDevices = normalizeKnownDevices(p.KnownDevices)
}

// normalizeKnownDevices drops unusable entries, collapses records that point at
// the same device, and trims the history to the most recently used or seen.
func normalizeKnownDevices(devices []KnownDevice) []KnownDevice {
	if len(devices) == 0 {
		return nil
	}
	cleaned := make([]KnownDevice, 0, len(devices))
	for _, device := range devices {
		device.BaseURL = strings.TrimSpace(device.BaseURL)
		if device.BaseURL == "" {
			continue
		}
		if strings.TrimSpace(device.Name) == "" {
			device.Name = hostLabel(device.BaseURL)
		}
		cleaned = append(cleaned, device)
	}
	// Merge into the most recent record, so the surviving entry is the one the
	// user last reached the device through.
	slices.SortStableFunc(cleaned, func(a, b KnownDevice) int {
		return b.recency().Compare(a.recency())
	})

	out := make([]KnownDevice, 0, len(cleaned))
	at := make(map[string]int, len(cleaned)*2)
	for _, device := range cleaned {
		index := -1
		for _, key := range deviceIdentityKeys(device.BaseURL, device.Host, device.IP) {
			if found, ok := at[key]; ok {
				index = found
				break
			}
		}
		if index < 0 {
			out = append(out, device)
			index = len(out) - 1
		} else {
			out[index] = mergeKnownDevices(out[index], device)
		}
		for _, key := range deviceIdentityKeys(out[index].BaseURL, out[index].Host, out[index].IP) {
			at[key] = index
		}
	}
	if len(out) > maxKnownDevices {
		out = out[:maxKnownDevices]
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// deviceIdentityKeys returns every address the device answers to. Two records
// describe the same device when any key matches, which is what collapses the
// http:// and https:// entries a scan can produce for one box.
func deviceIdentityKeys(baseURL, host, ip string) []string {
	keys := make([]string, 0, 3)
	for _, candidate := range []string{hostLabel(baseURL), host, ip} {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate == "" || slices.Contains(keys, candidate) {
			continue
		}
		keys = append(keys, candidate)
	}
	return keys
}

// mergeKnownDevices folds an older duplicate into the newer record, keeping
// whichever facts each one knows.
func mergeKnownDevices(primary, other KnownDevice) KnownDevice {
	if other.LastSeen.After(primary.LastSeen) {
		primary.LastSeen = other.LastSeen
	}
	if other.LastUsed.After(primary.LastUsed) {
		primary.LastUsed = other.LastUsed
	}
	if primary.Host == "" {
		primary.Host = other.Host
	}
	if primary.IP == "" {
		primary.IP = other.IP
	}
	if primary.Scheme == "" {
		primary.Scheme = other.Scheme
	}
	// A resolved hostname beats a bare address as a display name.
	if isAddressLabel(primary.Name) && !isAddressLabel(other.Name) {
		primary.Name = other.Name
	}
	return primary
}

// isAddressLabel reports whether a name is just an address, meaning we never
// learned what the device calls itself.
func isAddressLabel(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return true
	}
	if _, err := netip.ParseAddr(name); err == nil {
		return true
	}
	_, err := netip.ParseAddrPort(name)
	return err == nil
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
