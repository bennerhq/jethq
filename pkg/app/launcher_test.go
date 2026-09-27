package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bennerhq/jethq/pkg/discovery"
	"github.com/bennerhq/jethq/pkg/ui"
)

func TestRememberedDevicesSurviveRestart(t *testing.T) {
	originalHomeDir := userHomeDir
	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHomeDir })

	app, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if len(app.discovered) != 0 {
		t.Fatalf("discovered = %d entries, want empty list on a fresh profile", len(app.discovered))
	}

	app.addDiscoveredDevice(discovery.Device{
		Name:      "jetkvm.local",
		BaseURL:   "http://192.168.1.50",
		Host:      "jetkvm.local",
		IP:        "192.168.1.50",
		Scheme:    "http",
		IsSetup:   true,
		UpdatedAt: time.Unix(1700000000, 0),
	})

	if len(app.discovered) != 1 || !app.discovered[0].Confirmed {
		t.Fatalf("discovered = %+v, want one confirmed entry", app.discovered)
	}

	data, err := os.ReadFile(filepath.Join(home, ".config", "jethq", "config.json"))
	if err != nil {
		t.Fatalf("reading preferences: %v", err)
	}
	var stored struct {
		KnownDevices []KnownDevice `json:"known_devices"`
	}
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("decoding preferences: %v", err)
	}
	if len(stored.KnownDevices) != 1 || stored.KnownDevices[0].BaseURL != "http://192.168.1.50" {
		t.Fatalf("known_devices = %+v, want the discovered device persisted", stored.KnownDevices)
	}

	restarted, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if len(restarted.discovered) != 1 {
		t.Fatalf("discovered = %+v, want the remembered device replayed at startup", restarted.discovered)
	}
	replayed := restarted.discovered[0]
	if replayed.BaseURL != "http://192.168.1.50" || replayed.Name != "jetkvm.local" {
		t.Fatalf("replayed device = %+v, want the saved name and base URL", replayed)
	}
	if replayed.Confirmed {
		t.Fatal("replayed device is marked confirmed, want unconfirmed until this session's scan answers")
	}

	restarted.addDiscoveredDevice(discovery.Device{
		Name:      "jetkvm.local",
		BaseURL:   "http://192.168.1.50",
		IsSetup:   true,
		UpdatedAt: time.Unix(1700000900, 0),
	})
	if len(restarted.discovered) != 1 {
		t.Fatalf("discovered = %+v, want the scan to update the remembered entry in place", restarted.discovered)
	}
	if !restarted.discovered[0].Confirmed {
		t.Fatal("remembered device stayed unconfirmed after the scan answered")
	}
}

func TestSelectedDeviceTracksBaseURLAcrossRescans(t *testing.T) {
	// Scans persist devices, so keep this off the real user profile.
	originalHomeDir := userHomeDir
	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHomeDir })

	app, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	app.discovered = []launcherDevice{
		{Device: discovery.Device{Name: "alpha", BaseURL: "http://10.0.0.1"}, Confirmed: true},
		{Device: discovery.Device{Name: "bravo", BaseURL: "http://10.0.0.2"}, Confirmed: true},
	}

	if _, ok := app.selectedDevice(); ok {
		t.Fatal("expected no selection before an arrow key is pressed")
	}

	app.launcherSelection = "http://10.0.0.2"
	if got := app.selectedDeviceIndex(); got != 1 {
		t.Fatalf("selectedDeviceIndex() = %d, want 1", got)
	}

	// A later scan sorts a new device in ahead of the selected one.
	app.addDiscoveredDevice(discovery.Device{Name: "aardvark", BaseURL: "http://10.0.0.3"})
	selected, ok := app.selectedDevice()
	if !ok || selected.BaseURL != "http://10.0.0.2" {
		t.Fatalf("selectedDevice() = %+v (ok=%v), want the selection to follow the device, not the index", selected, ok)
	}

	// A device that disappears from the list releases the selection.
	app.discovered = app.discovered[:1]
	app.syncLauncherSelection()
	if app.launcherSelection != "" {
		t.Fatalf("launcherSelection = %q, want it cleared when the device leaves the list", app.launcherSelection)
	}
}

func TestLastUsedDeviceIsPreselectedOnStartup(t *testing.T) {
	originalHomeDir := userHomeDir
	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHomeDir })

	app, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	// Two devices on the network, neither used yet.
	app.addDiscoveredDevice(discovery.Device{Name: "alpha", BaseURL: "http://10.0.0.1", UpdatedAt: time.Unix(1700000000, 0)})
	app.addDiscoveredDevice(discovery.Device{Name: "bravo", BaseURL: "http://10.0.0.2", UpdatedAt: time.Unix(1700000001, 0)})
	if app.launcherSelection != "" {
		t.Fatalf("launcherSelection = %q, want no selection before anything has been used", app.launcherSelection)
	}

	app.connectTo("http://10.0.0.2")

	restarted, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if restarted.launcherSelection != "http://10.0.0.2" {
		t.Fatalf("launcherSelection = %q, want the last used device pre-selected", restarted.launcherSelection)
	}
	selected, ok := restarted.selectedDevice()
	if !ok || selected.BaseURL != "http://10.0.0.2" {
		t.Fatalf("selectedDevice() = %+v (ok=%v), want the last used device highlighted in the list", selected, ok)
	}
	if selected.Name != "bravo" {
		t.Fatalf("selected.Name = %q, want the scanned name preserved", selected.Name)
	}

	// A later scan must not erase the fact that the device was connected to.
	restarted.addDiscoveredDevice(discovery.Device{Name: "bravo", BaseURL: "http://10.0.0.2", UpdatedAt: time.Unix(1700009999, 0)})
	if got := lastUsedDevice(restarted.prefs.KnownDevices); got != "http://10.0.0.2" {
		t.Fatalf("lastUsedDevice() = %q after a rescan, want it preserved", got)
	}
}

func TestConnectingToTypedHostRemembersIt(t *testing.T) {
	originalHomeDir := userHomeDir
	home := t.TempDir()
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHomeDir })

	app, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	app.connectFromLauncher("jetkvm.local")

	if len(app.discovered) != 1 || app.discovered[0].BaseURL != "http://jetkvm.local" {
		t.Fatalf("discovered = %+v, want the typed host added to the list", app.discovered)
	}
	if app.discovered[0].Name != "jetkvm.local" {
		t.Fatalf("name = %q, want the host without its scheme", app.discovered[0].Name)
	}
	if app.discovered[0].Confirmed {
		t.Fatal("typed host is marked confirmed, want unconfirmed until a scan answers for it")
	}

	restarted, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	if restarted.launcherSelection != "http://jetkvm.local" {
		t.Fatalf("launcherSelection = %q, want the typed host pre-selected next launch", restarted.launcherSelection)
	}
}

func TestLauncherVisibleRangeFollowsSelection(t *testing.T) {
	tests := []struct {
		name            string
		count           int
		selection       int
		wantStart       int
		wantEnd         int
		wantSelectionIn bool
	}{
		{name: "short list shows everything", count: 3, selection: -1, wantStart: 0, wantEnd: 3},
		{name: "no selection shows the top", count: 12, selection: -1, wantStart: 0, wantEnd: launcherVisibleDevices},
		{name: "selection inside the first page", count: 12, selection: 2, wantStart: 0, wantEnd: launcherVisibleDevices, wantSelectionIn: true},
		{name: "selection scrolls the window", count: 12, selection: 8, wantStart: 2, wantEnd: 9, wantSelectionIn: true},
		{name: "last device pins the window to the end", count: 12, selection: 11, wantStart: 5, wantEnd: 12, wantSelectionIn: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := launcherVisibleRange(tt.count, tt.selection)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Fatalf("launcherVisibleRange(%d, %d) = (%d, %d), want (%d, %d)", tt.count, tt.selection, start, end, tt.wantStart, tt.wantEnd)
			}
			if tt.wantSelectionIn && (tt.selection < start || tt.selection >= end) {
				t.Fatalf("selection %d fell outside the visible window [%d, %d)", tt.selection, start, end)
			}
			if end > tt.count {
				t.Fatalf("end %d exceeds device count %d", end, tt.count)
			}
		})
	}
}

func TestLauncherDeviceStyling(t *testing.T) {
	for _, theme := range []ui.Theme{ui.DarkTheme(), ui.LightTheme()} {
		ctx := &ui.Context{Theme: theme}
		remembered := launcherDevice{Device: discovery.Device{BaseURL: "http://10.0.0.1"}}
		confirmed := launcherDevice{Device: discovery.Device{BaseURL: "http://10.0.0.1", IsSetup: true}, Confirmed: true}

		if got := (launcherDevicePanelElement{device: remembered}).panel(ctx).Stroke; got != theme.WarningStroke {
			t.Fatalf("remembered stroke = %v, want the warning colour %v", got, theme.WarningStroke)
		}
		if got := (launcherDevicePanelElement{device: confirmed}).panel(ctx).Stroke; got != theme.ActiveStroke {
			t.Fatalf("confirmed stroke = %v, want the regular active colour %v", got, theme.ActiveStroke)
		}

		unselected := (launcherDevicePanelElement{device: confirmed}).panel(ctx).Fill
		selected := (launcherDevicePanelElement{device: confirmed, selected: true}).panel(ctx).Fill
		if unselected != theme.SectionFill {
			t.Fatalf("unselected fill = %v, want %v", unselected, theme.SectionFill)
		}
		if selected != theme.SelectionFill {
			t.Fatalf("selected fill = %v, want %v", selected, theme.SelectionFill)
		}

		state, stateColor, _ := (launcherDeviceRowElement{device: remembered}).launcherDeviceState(ctx)
		if state != "Remembered" || stateColor != theme.WarningStroke {
			t.Fatalf("remembered state = (%q, %v), want (\"Remembered\", %v)", state, stateColor, theme.WarningStroke)
		}
		state, stateColor, _ = (launcherDeviceRowElement{device: confirmed}).launcherDeviceState(ctx)
		if state != "Configured" || stateColor != theme.AccentText {
			t.Fatalf("confirmed state = (%q, %v), want (\"Configured\", %v)", state, stateColor, theme.AccentText)
		}
	}
}

func TestNormalizeKnownDevicesCollapsesSchemeDuplicates(t *testing.T) {
	// Exactly the shape a real profile picked up: one device recorded twice
	// because a sweep reached it over http and a later one over https.
	got := normalizeKnownDevices([]KnownDevice{
		{Name: "192.168.86.29", BaseURL: "http://192.168.86.29", IP: "192.168.86.29", Scheme: "http", IsSetup: true,
			LastSeen: time.Unix(2000, 0), LastUsed: time.Unix(2100, 0)},
		{Name: "jetkvm.lan", BaseURL: "https://192.168.86.29", IP: "192.168.86.29", Scheme: "https", IsSetup: true,
			LastSeen: time.Unix(3000, 0)},
	})

	if len(got) != 1 {
		t.Fatalf("normalizeKnownDevices() = %+v, want the two schemes collapsed into one device", got)
	}
	device := got[0]
	if device.BaseURL != "https://192.168.86.29" {
		t.Fatalf("BaseURL = %q, want the most recently seen record to survive", device.BaseURL)
	}
	if device.Name != "jetkvm.lan" {
		t.Fatalf("Name = %q, want the resolved hostname rather than the address", device.Name)
	}
	if !device.LastSeen.Equal(time.Unix(3000, 0)) {
		t.Fatalf("LastSeen = %v, want the newer sighting", device.LastSeen)
	}
	if !device.LastUsed.Equal(time.Unix(2100, 0)) {
		t.Fatalf("LastUsed = %v, want the connection recorded against the other scheme", device.LastUsed)
	}
}

func TestScanReplacesDuplicateRowAndKeepsSelection(t *testing.T) {
	app, err := New(Config{})
	if err != nil {
		t.Fatal(err)
	}
	app.discovered = []launcherDevice{
		{Device: discovery.Device{Name: "192.168.86.29", BaseURL: "https://192.168.86.29", IP: "192.168.86.29"}},
	}
	app.launcherSelection = "https://192.168.86.29"

	app.addDiscoveredDevice(discovery.Device{
		Name: "jetkvm.lan", BaseURL: "http://192.168.86.29", Host: "jetkvm.lan", IP: "192.168.86.29",
		Scheme: "http", IsSetup: true, UpdatedAt: time.Unix(1700000000, 0),
	})

	if len(app.discovered) != 1 {
		t.Fatalf("discovered = %+v, want one row for one device", app.discovered)
	}
	if app.launcherSelection != "http://192.168.86.29" {
		t.Fatalf("launcherSelection = %q, want it to follow the device to its new base URL", app.launcherSelection)
	}
	if selected, ok := app.selectedDevice(); !ok || !selected.Confirmed {
		t.Fatalf("selectedDevice() = %+v (ok=%v), want the confirmed row still selected", selected, ok)
	}
}

func TestDeviceTitleAndSubtitlePairNameWithAddress(t *testing.T) {
	tests := []struct {
		name         string
		device       discovery.Device
		wantTitle    string
		wantSubtitle string
	}{
		{
			name:         "resolved hostname shows above the address",
			device:       discovery.Device{Name: "jetkvm.lan", Host: "jetkvm.lan", IP: "192.168.86.29", BaseURL: "http://192.168.86.29", Scheme: "http"},
			wantTitle:    "jetkvm.lan",
			wantSubtitle: "192.168.86.29",
		},
		{
			name:         "https is called out next to the address",
			device:       discovery.Device{Name: "jetkvm.lan", IP: "192.168.86.29", BaseURL: "https://192.168.86.29", Scheme: "https"},
			wantTitle:    "jetkvm.lan",
			wantSubtitle: "192.168.86.29 (https)",
		},
		{
			name:         "unnamed device falls back to its address",
			device:       discovery.Device{Name: "192.168.86.29", IP: "192.168.86.29", BaseURL: "http://192.168.86.29", Scheme: "http"},
			wantTitle:    "192.168.86.29",
			wantSubtitle: "http://192.168.86.29",
		},
		{
			// The name is the address here, so the URL is the only thing the
			// second line can usefully add.
			name:         "hand typed host keeps its name",
			device:       discovery.Device{Name: "jetkvm.local", BaseURL: "http://jetkvm.local"},
			wantTitle:    "jetkvm.local",
			wantSubtitle: "http://jetkvm.local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deviceTitle(tt.device); got != tt.wantTitle {
				t.Fatalf("deviceTitle() = %q, want %q", got, tt.wantTitle)
			}
			if got := deviceSubtitle(tt.device); got != tt.wantSubtitle {
				t.Fatalf("deviceSubtitle() = %q, want %q", got, tt.wantSubtitle)
			}
		})
	}
}

func TestNormalizeKnownDevicesTrimsHistory(t *testing.T) {
	devices := []KnownDevice{
		{BaseURL: "  ", Name: "blank"},
		{BaseURL: "http://10.0.0.2", LastSeen: time.Unix(200, 0)},
		{BaseURL: "http://10.0.0.2", Name: "duplicate", LastSeen: time.Unix(100, 0)},
		{BaseURL: "http://10.0.0.3", LastSeen: time.Unix(300, 0)},
	}
	for i := 0; i < maxKnownDevices; i++ {
		devices = append(devices, KnownDevice{BaseURL: "http://10.1.0." + string(rune('a'+i))})
	}

	got := normalizeKnownDevices(devices)
	if len(got) != maxKnownDevices {
		t.Fatalf("len = %d, want the history capped at %d", len(got), maxKnownDevices)
	}
	if got[0].BaseURL != "http://10.0.0.3" || got[1].BaseURL != "http://10.0.0.2" {
		t.Fatalf("got[0..1] = %+v, want most recently seen devices first", got[:2])
	}
	if got[1].Name != "duplicate" {
		t.Fatalf("got[1].Name = %q, want the real name merged in from the duplicate record", got[1].Name)
	}
	for _, device := range got {
		if device.Name == "blank" {
			t.Fatal("entry without a base URL survived normalization")
		}
	}
}
