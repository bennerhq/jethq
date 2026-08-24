package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type IconKind uint8

const (
	IconReconnect IconKind = iota
	IconMouse
	IconPaste
	IconMedia
	IconStats
	IconTerminal
	IconMinus
	IconPlus
	IconPower
	IconSettings
	IconFullscreen
	IconClose
)

type IconButton struct {
	Kind    IconKind
	Active  bool
	Enabled bool
	Alpha   float64
	Dock    bool
	Hovered bool
}

func (b IconButton) Measure(_ *Context, constraints Constraints) Size {
	return constraints.Clamp(Size{W: constraints.MaxW, H: constraints.MaxH})
}

func (b IconButton) Draw(ctx *Context, bounds Rect) {
	fill := withAlpha(ctx.Theme.ButtonFill, b.Alpha)
	stroke := withAlpha(ctx.Theme.ButtonStroke, b.Alpha)
	icon := withAlpha(ctx.Theme.ButtonText, b.Alpha)
	if b.Active {
		fill = withAlpha(ctx.Theme.ActiveFill, b.Alpha)
		stroke = withAlpha(ctx.Theme.ActiveStroke, b.Alpha)
		icon = rgba(244, 248, 252, 255, b.Alpha)
	}
	if !b.Enabled {
		fill = withAlpha(ctx.Theme.DisabledFill, b.Alpha)
		stroke = withAlpha(ctx.Theme.ButtonStroke, b.Alpha*0.75)
		icon = withAlpha(ctx.Theme.DisabledText, b.Alpha)
	}
	if b.Hovered && b.Enabled {
		fill = withAlpha(ctx.Theme.ActiveFill, b.Alpha*0.9)
		stroke = withAlpha(ctx.Theme.ActiveStroke, b.Alpha)
		icon = rgba(244, 248, 252, 255, b.Alpha)
	}
	if b.Dock && b.Enabled {
		icon = withAlpha(ctx.Theme.Title, b.Alpha)
	}
	if b.Dock {
		// The surrounding chrome strip supplies the shared glass-like surface.
		// Only selected controls get their own rounded highlight.
		if b.Active || b.Hovered {
			ctx.FillStrokedRoundedRect(bounds.Inset(UniformInsets(3)), 1, 7, stroke, fill)
		}
	} else {
		ctx.FillRect(bounds, fill)
		ctx.StrokeRect(bounds, 1, stroke)
	}
	drawIcon(ctx, b.Kind, bounds, icon, b.Active)
}

type Tooltip struct {
	Text  string
	Alpha float64
}

func (t Tooltip) Measure(_ *Context, constraints Constraints) Size {
	return constraints.Clamp(Size{W: constraints.MaxW, H: 28})
}

func (t Tooltip) Draw(ctx *Context, bounds Rect) {
	ctx.FillRect(bounds, withAlpha(ctx.Theme.ModalFill, t.Alpha))
	ctx.StrokeRect(bounds, 1, withAlpha(ctx.Theme.ModalStroke, t.Alpha))
	DrawText(ctx.Screen, t.Text, bounds.X+10, bounds.Y+8, 13, withAlpha(ctx.Theme.Body, t.Alpha))
}

func drawIcon(ctx *Context, kind IconKind, r Rect, clr color.Color, active bool) {
	cx := r.X + r.W/2
	cy := r.Y + r.H/2
	// Keep the control hit area generous, but use a restrained 26px glyph
	// canvas so the symbols read like a compact native toolbar rather than
	// oversized illustrations.
	left := r.X + 12
	right := r.X + r.W - 12
	top := r.Y + 12
	bottom := r.Y + r.H - 12
	if phosphor := phosphorIcon(kind); phosphor != nil {
		glyphSize := right - left
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(glyphSize/phosphorRasterSize, glyphSize/phosphorRasterSize)
		op.GeoM.Translate(left, top)
		op.ColorScale.ScaleWithColor(clr)
		ctx.Screen.DrawImage(phosphor, op)
		return
	}
	mid := r.Y + r.H/2
	switch kind {
	case IconReconnect:
		// Refresh-cw: a lighter, more familiar reconnect mark.
		ctx.StrokeLine(Point{left + 4, top + 5}, Point{right - 3, top + 5}, 1.4, clr)
		ctx.StrokeLine(Point{right - 3, top + 5}, Point{right - 7, top + 1}, 1.4, clr)
		ctx.StrokeLine(Point{right - 3, top + 5}, Point{right - 7, top + 9}, 1.4, clr)
		ctx.StrokeLine(Point{right - 4, bottom - 5}, Point{left + 3, bottom - 5}, 1.4, clr)
		ctx.StrokeLine(Point{left + 3, bottom - 5}, Point{left + 7, bottom - 9}, 1.4, clr)
		ctx.StrokeLine(Point{left + 3, bottom - 5}, Point{left + 7, bottom - 1}, 1.4, clr)
	case IconMouse:
		if active {
			ctx.StrokeLine(Point{cx, top}, Point{cx, bottom}, 1.5, clr)
			ctx.StrokeLine(Point{left, cy}, Point{right, cy}, 1.5, clr)
			ctx.StrokeLine(Point{cx, top}, Point{cx - 3, top + 3}, 1.5, clr)
			ctx.StrokeLine(Point{cx, top}, Point{cx + 3, top + 3}, 1.5, clr)
			ctx.StrokeLine(Point{cx, bottom}, Point{cx - 3, bottom - 3}, 1.5, clr)
			ctx.StrokeLine(Point{cx, bottom}, Point{cx + 3, bottom - 3}, 1.5, clr)
		} else {
			ctx.StrokeLine(Point{left + 2, top}, Point{left + 2, bottom - 1}, 1.5, clr)
			ctx.StrokeLine(Point{left + 2, top}, Point{right - 1, cy}, 1.5, clr)
			ctx.StrokeLine(Point{left + 2, top}, Point{cx + 1, bottom - 2}, 1.5, clr)
		}
	case IconPaste:
		ctx.StrokeRect(Rect{X: left + 1, Y: top + 3, W: right - left - 2, H: bottom - top - 3}, 1.4, clr)
		ctx.StrokeLine(Point{left + 4, top + 7}, Point{right - 4, top + 7}, 1.4, clr)
		ctx.StrokeLine(Point{cx, top + 7}, Point{cx, top + 2}, 1.4, clr)
	case IconMedia:
		// Hard-drive, matching the virtual-media action rather than a generic box.
		ctx.StrokeRect(Rect{X: left, Y: top + 4, W: right - left, H: bottom - top - 6}, 1.4, clr)
		ctx.StrokeLine(Point{left + 2, cy + 3}, Point{right - 2, cy + 3}, 1.4, clr)
		ctx.FillCircle(Point{X: left + 5, Y: bottom - 5}, 1.2, clr)
		ctx.FillCircle(Point{X: left + 9, Y: bottom - 5}, 1.2, clr)
	case IconStats:
		ctx.StrokeLine(Point{left + 3, bottom - 1}, Point{left + 3, mid + 4}, 1.8, clr)
		ctx.StrokeLine(Point{cx, bottom - 1}, Point{cx, top + 5}, 1.8, clr)
		ctx.StrokeLine(Point{right - 3, bottom - 1}, Point{right - 3, mid - 1}, 1.8, clr)
	case IconTerminal:
		ctx.StrokeRect(Rect{X: left, Y: top + 1, W: right - left, H: bottom - top - 2}, 1.4, clr)
		ctx.StrokeLine(Point{left + 4, top + 6}, Point{left + 7, top + 9}, 1.4, clr)
		ctx.StrokeLine(Point{left + 4, bottom - 7}, Point{left + 7, top + 9}, 1.4, clr)
		ctx.StrokeLine(Point{left + 10, bottom - 6}, Point{right - 4, bottom - 6}, 1.4, clr)
	case IconMinus:
		ctx.StrokeLine(Point{left, cy}, Point{right, cy}, 2, clr)
	case IconPlus:
		ctx.StrokeLine(Point{left, cy}, Point{right, cy}, 2, clr)
		ctx.StrokeLine(Point{cx, top}, Point{cx, bottom}, 2, clr)
	case IconPower:
		ctx.StrokeLine(Point{cx, top - 1}, Point{cx, cy - 2}, 2, clr)
		ctx.StrokeLine(Point{left + 3, top + 4}, Point{left, mid}, 1.5, clr)
		ctx.StrokeLine(Point{left, mid}, Point{left + 4, bottom - 1}, 1.5, clr)
		ctx.StrokeLine(Point{left + 4, bottom - 1}, Point{right - 4, bottom - 1}, 1.5, clr)
		ctx.StrokeLine(Point{right - 4, bottom - 1}, Point{right, mid}, 1.5, clr)
		ctx.StrokeLine(Point{right, mid}, Point{right - 3, top + 4}, 1.5, clr)
	case IconSettings:
		ctx.StrokeLine(Point{left, top + 3}, Point{right, top + 3}, 1.4, clr)
		ctx.StrokeLine(Point{left, cy}, Point{right, cy}, 1.4, clr)
		ctx.StrokeLine(Point{left, bottom - 3}, Point{right, bottom - 3}, 1.4, clr)
		ctx.FillCircle(Point{X: cx - 4, Y: top + 3}, 2.2, clr)
		ctx.FillCircle(Point{X: cx + 4, Y: cy}, 2.2, clr)
		ctx.FillCircle(Point{X: cx - 1, Y: bottom - 3}, 2.2, clr)
	case IconFullscreen:
		ctx.StrokeLine(Point{left, top + 4}, Point{left, top}, 1.6, clr)
		ctx.StrokeLine(Point{left, top}, Point{left + 4, top}, 1.6, clr)
		ctx.StrokeLine(Point{right, top + 4}, Point{right, top}, 1.6, clr)
		ctx.StrokeLine(Point{right - 4, top}, Point{right, top}, 1.6, clr)
		ctx.StrokeLine(Point{left, bottom - 4}, Point{left, bottom}, 1.6, clr)
		ctx.StrokeLine(Point{left, bottom}, Point{left + 4, bottom}, 1.6, clr)
		ctx.StrokeLine(Point{right, bottom - 4}, Point{right, bottom}, 1.6, clr)
		ctx.StrokeLine(Point{right - 4, bottom}, Point{right, bottom}, 1.6, clr)
	case IconClose:
		ctx.StrokeLine(Point{left, top}, Point{right, bottom}, 1.8, clr)
		ctx.StrokeLine(Point{right, top}, Point{left, bottom}, 1.8, clr)
	}
}

func rgba(r, g, b, a uint8, alpha float64) color.Color {
	if alpha <= 0 {
		return color.RGBA{}
	}
	if alpha > 1 {
		alpha = 1
	}
	return color.RGBA{
		R: r,
		G: g,
		B: b,
		A: uint8(float64(a) * alpha),
	}
}

func withAlpha(clr color.Color, alpha float64) color.Color {
	return WithAlpha(clr, alpha)
}

// WithAlpha returns clr with its existing opacity scaled by alpha.
func WithAlpha(clr color.Color, alpha float64) color.Color {
	if alpha <= 0 {
		return color.RGBA{}
	}
	if alpha >= 1 {
		return clr
	}
	r, g, b, a := clr.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8((float64(a>>8) * alpha)),
	}
}
