package app

import (
	"fmt"
	"image/color"
	"slices"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/bennerhq/jethq/pkg/discovery"
	"github.com/bennerhq/jethq/pkg/ui"
)

func (a *App) syncDiscovery() {
	if a.discovery == nil {
		return
	}
	for {
		select {
		case device := <-a.discovery.Updates():
			a.addDiscoveredDevice(device)
		default:
			return
		}
	}
}

// launcherDevice is a launcher list entry. Confirmed separates devices this
// session's scan has actually answered for from the ones merely replayed out of
// preferences, which the list renders in a distinct colour until they answer.
type launcherDevice struct {
	discovery.Device
	Confirmed bool
}

// rememberedDevices replays the saved device history into launcher entries so
// the list is populated on the very first frame, before any subnet sweep lands.
func rememberedDevices(known []KnownDevice) []launcherDevice {
	if len(known) == 0 {
		return nil
	}
	out := make([]launcherDevice, 0, len(known))
	for _, device := range known {
		if strings.TrimSpace(device.BaseURL) == "" {
			continue
		}
		name := device.Name
		if strings.TrimSpace(name) == "" {
			name = device.BaseURL
		}
		out = append(out, launcherDevice{Device: discovery.Device{
			Name:      name,
			BaseURL:   device.BaseURL,
			Host:      device.Host,
			IP:        device.IP,
			Scheme:    device.Scheme,
			IsSetup:   device.IsSetup,
			UpdatedAt: device.LastSeen,
		}})
	}
	sortLauncherDevices(out)
	return out
}

// lastUsedDevice returns the base URL of the most recently connected device, so
// the launcher can pre-select it and Enter reconnects without any navigation.
func lastUsedDevice(known []KnownDevice) string {
	baseURL := ""
	var used time.Time
	for _, device := range known {
		if device.LastUsed.IsZero() || strings.TrimSpace(device.BaseURL) == "" {
			continue
		}
		if baseURL == "" || device.LastUsed.After(used) {
			baseURL = device.BaseURL
			used = device.LastUsed
		}
	}
	return baseURL
}

// rememberUsedDevice records a connection target. Devices typed by hand are
// added to the history too: they were used, so they belong in the list next
// time even though no scan has ever answered for them.
func (a *App) rememberUsedDevice(baseURL string) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return
	}
	a.launcherSelection = baseURL
	now := time.Now()
	for i := range a.prefs.KnownDevices {
		if a.prefs.KnownDevices[i].BaseURL == baseURL {
			a.prefs.KnownDevices[i].LastUsed = now
			a.prefs.KnownDevices = normalizeKnownDevices(a.prefs.KnownDevices)
			a.savePreferences()
			return
		}
	}
	a.prefs.KnownDevices = normalizeKnownDevices(append(a.prefs.KnownDevices, KnownDevice{
		Name:     hostLabel(baseURL),
		BaseURL:  baseURL,
		LastUsed: now,
	}))
	a.savePreferences()
	a.addRememberedDevice(baseURL)
}

// addRememberedDevice puts a hand-entered target into the launcher list right
// away so it is there if the user comes back from the password prompt.
func (a *App) addRememberedDevice(baseURL string) {
	if a.indexOfDevice(baseURL, "", "") >= 0 {
		return
	}
	a.discovered = append(a.discovered, launcherDevice{Device: discovery.Device{
		Name:    hostLabel(baseURL),
		BaseURL: baseURL,
	}})
	a.sortDiscovered()
}

func hostLabel(baseURL string) string {
	trimmed := strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
	if trimmed == "" {
		return baseURL
	}
	return trimmed
}

func (a *App) addDiscoveredDevice(device discovery.Device) {
	entry := launcherDevice{Device: device, Confirmed: true}
	if index := a.indexOfDevice(device.BaseURL, device.Host, device.IP); index >= 0 {
		// Replacing rather than appending keeps one row per device even when a
		// scan reaches it on a different scheme than last time.
		if a.launcherSelection == a.discovered[index].BaseURL {
			a.launcherSelection = device.BaseURL
		}
		a.discovered[index] = entry
	} else {
		a.discovered = append(a.discovered, entry)
	}
	a.sortDiscovered()
	a.rememberDevice(device)
}

// indexOfDevice finds the row describing the same device, matching on any
// address it is known by rather than on the exact base URL.
func (a *App) indexOfDevice(baseURL, host, ip string) int {
	keys := deviceIdentityKeys(baseURL, host, ip)
	for i := range a.discovered {
		for _, key := range deviceIdentityKeys(a.discovered[i].BaseURL, a.discovered[i].Host, a.discovered[i].IP) {
			if slices.Contains(keys, key) {
				return i
			}
		}
	}
	return -1
}

// rememberDevice records a freshly scanned device in preferences so it can be
// shown again on the next launch.
func (a *App) rememberDevice(device discovery.Device) {
	if strings.TrimSpace(device.BaseURL) == "" {
		return
	}
	seenAt := device.UpdatedAt
	if seenAt.IsZero() {
		seenAt = time.Now()
	}
	known := KnownDevice{
		Name:     device.Name,
		BaseURL:  device.BaseURL,
		Host:     device.Host,
		IP:       device.IP,
		Scheme:   device.Scheme,
		IsSetup:  device.IsSetup,
		LastSeen: seenAt,
	}
	for i := range a.prefs.KnownDevices {
		if a.prefs.KnownDevices[i].BaseURL == known.BaseURL {
			// A scan says nothing about when the device was last connected to.
			known.LastUsed = a.prefs.KnownDevices[i].LastUsed
			if a.prefs.KnownDevices[i] == known {
				return
			}
			a.prefs.KnownDevices[i] = known
			a.prefs.KnownDevices = normalizeKnownDevices(a.prefs.KnownDevices)
			a.savePreferences()
			return
		}
	}
	a.prefs.KnownDevices = normalizeKnownDevices(append(a.prefs.KnownDevices, known))
	a.savePreferences()
}

func (a *App) sortDiscovered() {
	sortLauncherDevices(a.discovered)
}

func sortLauncherDevices(devices []launcherDevice) {
	slices.SortFunc(devices, func(a, b launcherDevice) int {
		if a.Name != b.Name {
			return strings.Compare(a.Name, b.Name)
		}
		return strings.Compare(a.BaseURL, b.BaseURL)
	})
}

func (a *App) syncLauncherInput() {
	switch a.launcherMode {
	case launcherModeBrowse:
		if a.textInput.FieldID != "launcher_focus_input" {
			a.textInput.Sync(&ui.TextInputBinding{
				ID:       "launcher_focus_input",
				Value:    a.launcherInput,
				TextSize: 15,
			})
		}
	case launcherModePassword:
		if a.textInput.FieldID != "launcher_focus_password" {
			a.textInput.Sync(&ui.TextInputBinding{
				ID:           "launcher_focus_password",
				Value:        a.launcherPassword,
				DisplayValue: strings.Repeat("*", len([]rune(a.launcherPassword))),
				TextSize:     15,
			})
		}
	}
	if a.launcherMode == launcherModeBrowse {
		a.syncLauncherSelection()
	}
	a.syncFocusedTextInput()
	if launcherConfirmPressed() {
		if a.launcherMode == launcherModePassword {
			a.connectFromLauncher(a.pendingTarget)
		} else if selected, ok := a.selectedDevice(); ok {
			a.connectFromLauncher(selected.BaseURL)
		} else {
			a.connectFromLauncher(a.launcherInput)
		}
		return
	}
}

func launcherConfirmPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter)
}

// selectedDevice resolves the keyboard selection, which is tracked by base URL
// rather than by index because every scan result re-sorts the list.
func (a *App) selectedDevice() (launcherDevice, bool) {
	index := a.selectedDeviceIndex()
	if index < 0 {
		return launcherDevice{}, false
	}
	return a.discovered[index], true
}

func (a *App) selectedDeviceIndex() int {
	if a.launcherSelection == "" {
		return -1
	}
	for i := range a.discovered {
		if a.discovered[i].BaseURL == a.launcherSelection {
			return i
		}
	}
	return -1
}

// syncLauncherSelection moves the highlight with the arrow keys. Stepping up
// past the first device drops the selection so Enter goes back to whatever is
// typed in the connect field.
func (a *App) syncLauncherSelection() {
	if len(a.discovered) == 0 {
		a.launcherSelection = ""
		return
	}
	index := a.selectedDeviceIndex()
	if a.launcherSelection != "" && index < 0 {
		// The selected device dropped off the list; fall back to the field.
		a.launcherSelection = ""
	}
	switch {
	case launcherKeyRepeated(ebiten.KeyDown):
		if index < len(a.discovered)-1 {
			index++
		}
		a.launcherSelection = a.discovered[index].BaseURL
	case launcherKeyRepeated(ebiten.KeyUp):
		if index < 0 {
			// Wrap into the list from the field so Up alone is useful too.
			a.launcherSelection = a.discovered[len(a.discovered)-1].BaseURL
			return
		}
		if index == 0 {
			a.launcherSelection = ""
			return
		}
		a.launcherSelection = a.discovered[index-1].BaseURL
	}
}

const (
	launcherKeyRepeatDelay    = 30
	launcherKeyRepeatInterval = 4
)

// launcherKeyRepeated reports a press plus the auto-repeat that follows when the
// key is held, so holding an arrow key walks the list.
func launcherKeyRepeated(key ebiten.Key) bool {
	held := inpututil.KeyPressDuration(key)
	if held == 1 {
		return true
	}
	if held <= launcherKeyRepeatDelay {
		return false
	}
	return (held-launcherKeyRepeatDelay)%launcherKeyRepeatInterval == 0
}

func (a *App) drawLauncher(screen *ebiten.Image) {
	screen.Fill(a.currentTheme().Background)

	if a.launcherMode == launcherModePassword {
		a.drawPasswordPrompt(screen)
		return
	}

	a.drawUIRoot(screen, &a.launcherRuntime, func(chromeButton) {}, launcherScreenElement{app: a})
}

func (a *App) drawPasswordPrompt(screen *ebiten.Image) {
	a.drawUIRoot(screen, &a.launcherRuntime, func(chromeButton) {}, launcherPasswordScreenElement{app: a})
}

type launcherScreenElement struct {
	app *App
}

func (launcherScreenElement) Measure(_ *ui.Context, constraints ui.Constraints) ui.Size {
	return constraints.Clamp(ui.Size{W: constraints.MaxW, H: constraints.MaxH})
}

func (e launcherScreenElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	validInput := strings.TrimSpace(e.app.launcherInput) != "" && isValidConnectHost(strings.TrimSpace(e.app.launcherInput))
	children := []ui.Child{
		ui.Fixed(ui.Row{Children: []ui.Child{
			ui.Flex(ui.Label{Text: "JetKVM", Size: 30, Color: ctx.Theme.Title}, 1),
			ui.Fixed(ui.Button{Label: "Settings", Enabled: true, OnClick: func() { e.app.openSettingsOverlay() }}),
		}, Spacing: 12}),
		ui.Fixed(ui.Spacer{H: 12}),
		ui.Fixed(ui.Label{Text: "Available devices on your local network", Size: 15, Color: ctx.Theme.Muted}),
		ui.Fixed(ui.Spacer{H: 28}),
		ui.Flex(ui.Constrained{
			MinH: 120,
			MaxH: 520,
			Child: ui.Panel{
				Fill:   ctx.Theme.PanelFill,
				Stroke: ctx.Theme.PanelStroke,
				Insets: ui.UniformInsets(18),
				Child:  launcherListElement(e),
			},
		}, 1),
		ui.Fixed(ui.Spacer{H: 18}),
		ui.Fixed(ui.Column{
			Children: []ui.Child{
				ui.Fixed(ui.Label{Text: "Connect by host, DNS name, or IP", Size: 13, Color: ctx.Theme.Muted}),
				ui.Fixed(ui.Spacer{H: 8}),
				ui.Fixed(ui.Row{
					Children: []ui.Child{
						ui.Flex(launcherInputElement(e), 1),
						ui.Fixed(ui.Button{Label: "Connect", Enabled: validInput, OnClick: func() {
							e.app.connectFromLauncher(e.app.launcherInput)
						}}),
					},
					Spacing: 12,
				}),
			},
		}),
	}
	if e.app.launcherError != "" {
		children = append(children,
			ui.Fixed(ui.Spacer{H: 12}),
			ui.Fixed(ui.Paragraph{Text: e.app.launcherError, Size: 12, Color: ctx.Theme.Error}),
		)
	} else if strings.TrimSpace(e.app.launcherInput) != "" && !validInput {
		children = append(children,
			ui.Fixed(ui.Spacer{H: 12}),
			ui.Fixed(ui.Paragraph{Text: "Enter a valid hostname or IP address.", Size: 12, Color: ctx.Theme.Error}),
		)
	}
	ui.Inset{
		Insets: ui.Insets{Top: 42, Right: 48, Bottom: 44, Left: 48},
		Child: ui.Constrained{
			MaxW:  1380,
			Child: ui.Column{Children: children},
		},
	}.Draw(ctx, bounds)
}

type launcherPasswordScreenElement struct {
	app *App
}

func (launcherPasswordScreenElement) Measure(_ *ui.Context, constraints ui.Constraints) ui.Size {
	return constraints.Clamp(ui.Size{W: constraints.MaxW, H: constraints.MaxH})
}

func (e launcherPasswordScreenElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	targetLabel := e.app.pendingTarget
	if targetLabel == "" {
		targetLabel = e.app.launcherInput
	}
	ui.Inset{
		Insets: ui.UniformInsets(48),
		Child: ui.Align{
			Horizontal: ui.AlignCenter,
			Vertical:   ui.AlignCenter,
			Child: ui.Column{
				Children: []ui.Child{
					ui.Fixed(ui.Label{Text: "JetKVM", Size: 30, Color: ctx.Theme.Title}),
					ui.Fixed(ui.Spacer{H: 26}),
					ui.Fixed(ui.Constrained{
						MaxW: 620,
						Child: ui.Panel{
							Fill:   ctx.Theme.PanelFill,
							Stroke: ctx.Theme.PanelStroke,
							Insets: ui.UniformInsets(24),
							Child: launcherPasswordElement{
								app:         e.app,
								targetLabel: targetLabel,
							},
						},
					}),
				},
			},
		},
	}.Draw(ctx, bounds)
}

type launcherListElement struct {
	app *App
}

func (e launcherListElement) Measure(_ *ui.Context, constraints ui.Constraints) ui.Size {
	return constraints.Clamp(ui.Size{W: constraints.MaxW, H: constraints.MaxH})
}

func (e launcherListElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	if len(e.app.discovered) == 0 {
		ui.Column{
			Children: []ui.Child{
				ui.Fixed(ui.Label{Text: "Scanning local subnets for JetKVM devices...", Size: 16, Color: ctx.Theme.AccentText}),
				ui.Fixed(ui.Spacer{H: 12}),
				ui.Fixed(ui.Paragraph{
					Text:  "Devices will appear here as soon as they answer the JetKVM HTTP status endpoint.",
					Size:  13,
					Color: ctx.Theme.Muted,
				}),
			},
		}.Draw(ctx, bounds)
		return
	}
	start, end := launcherVisibleRange(len(e.app.discovered), e.app.selectedDeviceIndex())
	children := make([]ui.Child, 0, (end-start)*2+2)
	var lastConfirmed time.Time
	for i := start; i < end; i++ {
		device := e.app.discovered[i]
		if i > start {
			children = append(children, ui.Fixed(ui.Spacer{H: 8}))
		}
		children = append(children, ui.Fixed(launcherDevicePanelElement{
			app:      e.app,
			device:   device,
			selected: device.BaseURL == e.app.launcherSelection,
		}))
		if device.Confirmed && device.UpdatedAt.After(lastConfirmed) {
			lastConfirmed = device.UpdatedAt
		}
	}
	footer := "Scanning local subnets; remembered devices shown until they answer..."
	if !lastConfirmed.IsZero() {
		footer = fmt.Sprintf("Updated %s", humanDiscoveryAge(lastConfirmed))
	}
	children = append(children,
		ui.Flex(ui.Spacer{}, 1),
		ui.Fixed(ui.Label{Text: "Up/Down to select, Enter to connect", Size: 11, Color: ctx.Theme.DisabledText}),
		ui.Fixed(ui.Spacer{H: 4}),
		ui.Fixed(ui.Label{Text: footer, Size: 11, Color: ctx.Theme.DisabledText}),
	)
	ui.Column{Children: children}.Draw(ctx, bounds)
}

// launcherVisibleDevices caps how many rows the list shows at once; the window
// slides so the keyboard selection stays on screen.
const launcherVisibleDevices = 7

func launcherVisibleRange(count, selection int) (int, int) {
	if count <= launcherVisibleDevices {
		return 0, count
	}
	start := 0
	if selection >= launcherVisibleDevices {
		start = selection - launcherVisibleDevices + 1
	}
	if start+launcherVisibleDevices > count {
		start = count - launcherVisibleDevices
	}
	return start, start + launcherVisibleDevices
}

type launcherDevicePanelElement struct {
	app      *App
	device   launcherDevice
	selected bool
}

func (e launcherDevicePanelElement) panel(ctx *ui.Context) ui.Panel {
	stroke := ctx.Theme.ActiveStroke
	if !e.device.Confirmed {
		stroke = ctx.Theme.WarningStroke
	}
	fill := ctx.Theme.SectionFill
	if e.selected {
		fill = ctx.Theme.SelectionFill
	}
	return ui.Panel{
		Fill:   fill,
		Stroke: stroke,
		Insets: ui.SymmetricInsets(16, 12),
		Child:  launcherDeviceRowElement{device: e.device},
	}
}

func (e launcherDevicePanelElement) Measure(ctx *ui.Context, constraints ui.Constraints) ui.Size {
	return e.panel(ctx).Measure(ctx, constraints)
}

func (e launcherDevicePanelElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	e.panel(ctx).Draw(ctx, bounds)
	if ctx.Runtime != nil {
		baseURL := e.device.BaseURL
		ctx.Runtime.Register(ui.Control{
			ID:      "discover:" + baseURL,
			Rect:    bounds,
			Enabled: true,
			OnClick: func(ui.PointerEvent) {
				e.app.connectFromLauncher(baseURL)
			},
		})
	} else {
		ctx.AddHit("discover:"+e.device.BaseURL, bounds, true)
	}
}

type launcherDeviceRowElement struct {
	device launcherDevice
}

// launcherDeviceState is the right-hand status label plus the colour the whole
// row is drawn in: remembered devices stay amber until this session's scan
// confirms them, at which point they switch to the regular accent colour.
func (e launcherDeviceRowElement) launcherDeviceState(ctx *ui.Context) (string, color.Color, color.Color) {
	if !e.device.Confirmed {
		return "Remembered", ctx.Theme.WarningStroke, ctx.Theme.WarningStroke
	}
	if !e.device.IsSetup {
		return "Needs setup", ctx.Theme.AccentText, ctx.Theme.Title
	}
	return "Configured", ctx.Theme.AccentText, ctx.Theme.Title
}

func (e launcherDeviceRowElement) row(ctx *ui.Context) ui.Row {
	state, stateColor, nameColor := e.launcherDeviceState(ctx)
	return ui.Row{
		AlignY: ui.AlignCenter,
		Children: []ui.Child{
			ui.Flex(ui.Column{
				Children: []ui.Child{
					ui.Fixed(ui.Label{Text: deviceTitle(e.device.Device), Size: 17, Color: nameColor}),
					ui.Fixed(ui.Spacer{H: 8}),
					ui.Fixed(ui.Label{Text: deviceSubtitle(e.device.Device), Size: 13, Color: ctx.Theme.Muted}),
				},
			}, 1),
			ui.Fixed(ui.Label{Text: state, Size: 13, Color: stateColor}),
		},
	}
}

// deviceTitle is what the device calls itself, falling back to its address when
// no name could be resolved.
func deviceTitle(device discovery.Device) string {
	for _, candidate := range []string{device.Name, device.Host} {
		if candidate = strings.TrimSpace(candidate); candidate != "" && !isAddressLabel(candidate) {
			return candidate
		}
	}
	return deviceAddress(device)
}

// deviceSubtitle pairs the name with the address it lives at. When the title is
// already the address there is nothing to add but the URL itself.
func deviceSubtitle(device discovery.Device) string {
	address := deviceAddress(device)
	if deviceTitle(device) == address {
		return device.BaseURL
	}
	scheme := strings.TrimSpace(device.Scheme)
	if scheme == "" || scheme == "http" {
		return address
	}
	return fmt.Sprintf("%s (%s)", address, scheme)
}

func deviceAddress(device discovery.Device) string {
	if ip := strings.TrimSpace(device.IP); ip != "" {
		return ip
	}
	return hostLabel(device.BaseURL)
}

func (e launcherDeviceRowElement) Measure(ctx *ui.Context, constraints ui.Constraints) ui.Size {
	return e.row(ctx).Measure(ctx, constraints)
}

func (e launcherDeviceRowElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	e.row(ctx).Draw(ctx, bounds)
}

type launcherInputElement struct {
	app *App
}

func (launcherInputElement) Measure(_ *ui.Context, constraints ui.Constraints) ui.Size {
	return constraints.Clamp(ui.Size{W: constraints.MaxW, H: 40})
}

func (e launcherInputElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	e.app.decorateTextField(ui.TextField{
		ID:               "launcher_focus_input",
		Value:            e.app.launcherInput,
		Placeholder:      "jetkvm.local or 192.168.1.50",
		Enabled:          true,
		TextSize:         15,
		FillColor:        ctx.Theme.InputFill,
		StrokeColor:      ctx.Theme.InputStroke,
		FocusColor:       ctx.Theme.InputFocus,
		TextColor:        ctx.Theme.Body,
		PlaceholderColor: ctx.Theme.DisabledText,
		CaretColor:       ctx.Theme.AccentText,
	}).Draw(ctx, bounds)
}

type launcherPasswordElement struct {
	app         *App
	targetLabel string
}

func (e launcherPasswordElement) Measure(ctx *ui.Context, constraints ui.Constraints) ui.Size {
	return ui.Column{Children: e.children()}.Measure(ctx, constraints)
}

func (e launcherPasswordElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	ui.Column{Children: e.children()}.Draw(ctx, bounds)
}

func (e launcherPasswordElement) children() []ui.Child {
	passDisplay := strings.Repeat("*", len([]rune(e.app.launcherPassword)))
	children := []ui.Child{
		ui.Fixed(ui.Label{Text: "Password Required", Size: 24, Color: e.app.currentTheme().Title}),
		ui.Fixed(ui.Spacer{H: 12}),
		ui.Fixed(ui.Paragraph{Text: e.targetLabel, Size: 14, Color: e.app.currentTheme().Muted}),
		ui.Fixed(ui.Spacer{H: 22}),
		ui.Fixed(launcherPasswordFieldElement{app: e.app, passDisplay: passDisplay}),
		ui.Fixed(ui.Spacer{H: 18}),
	}
	if e.app.launcherError != "" {
		children = append(children,
			ui.Fixed(ui.Paragraph{Text: e.app.launcherError, Size: 12, Color: e.app.currentTheme().Error}),
			ui.Fixed(ui.Spacer{H: 12}),
		)
	}
	children = append(children, ui.Fixed(ui.Row{
		Children: []ui.Child{
			ui.Fixed(ui.Button{Label: "Back", Enabled: true, OnClick: func() {
				e.app.launcherMode = launcherModeBrowse
				e.app.launcherPassword = ""
				e.app.launcherError = ""
				if e.app.cfg.BaseURL != "" && e.app.ctrl != nil {
					e.app.ctrl.Stop()
					e.app.ctrl = nil
				}
			}}),
			ui.Flex(ui.Spacer{}, 1),
			ui.Fixed(ui.Button{Label: "Connect", Enabled: strings.TrimSpace(e.app.launcherPassword) != "", OnClick: func() {
				e.app.connectFromLauncher(e.app.pendingTarget)
			}}),
		},
		Spacing: 12,
	}))
	return children
}

type launcherPasswordFieldElement struct {
	app         *App
	passDisplay string
}

func (launcherPasswordFieldElement) Measure(_ *ui.Context, constraints ui.Constraints) ui.Size {
	return constraints.Clamp(ui.Size{W: constraints.MaxW, H: 38})
}

func (e launcherPasswordFieldElement) Draw(ctx *ui.Context, bounds ui.Rect) {
	e.app.decorateTextField(ui.TextField{
		ID:               "launcher_focus_password",
		DisplayValue:     e.passDisplay,
		Placeholder:      "Local password",
		Enabled:          true,
		TextSize:         15,
		FillColor:        ctx.Theme.InputFill,
		StrokeColor:      ctx.Theme.WarningStroke,
		FocusColor:       ctx.Theme.WarningStroke,
		TextColor:        ctx.Theme.Body,
		PlaceholderColor: ctx.Theme.DisabledText,
		CaretColor:       ctx.Theme.Error,
	}).Draw(ctx, bounds)
}

func humanDiscoveryAge(at time.Time) string {
	if at.IsZero() {
		return "just now"
	}
	age := time.Since(at)
	switch {
	case age < time.Second:
		return "just now"
	case age < time.Minute:
		return fmt.Sprintf("%ds ago", int(age.Seconds()))
	default:
		return fmt.Sprintf("%dm ago", int(age.Minutes()))
	}
}
