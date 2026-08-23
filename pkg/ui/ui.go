package ui

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Point struct {
	X float64
	Y float64
}

type Size struct {
	W float64
	H float64
}

type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

func (r Rect) Contains(point Point) bool {
	return point.X >= r.X && point.X <= r.Right() && point.Y >= r.Y && point.Y <= r.Bottom()
}

func (r Rect) Inset(in Insets) Rect {
	x := r.X + in.Left
	y := r.Y + in.Top
	w := r.W - in.Left - in.Right
	h := r.H - in.Top - in.Bottom
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return Rect{X: x, Y: y, W: w, H: h}
}

func (r Rect) InnerSize(in Insets) Size {
	inner := r.Inset(in)
	return Size{W: inner.W, H: inner.H}
}

func (r Rect) Bottom() float64 {
	return r.Y + r.H
}

func (r Rect) Right() float64 {
	return r.X + r.W
}

type Insets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

func UniformInsets(v float64) Insets {
	return Insets{Top: v, Right: v, Bottom: v, Left: v}
}

func SymmetricInsets(horizontal, vertical float64) Insets {
	return Insets{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

type Constraints struct {
	MinW float64
	MaxW float64
	MinH float64
	MaxH float64
}

func NewConstraints(maxW, maxH float64) Constraints {
	return Constraints{MaxW: maxW, MaxH: maxH}
}

func (c Constraints) Tighten(size Size) Constraints {
	return Constraints{
		MinW: size.W,
		MaxW: size.W,
		MinH: size.H,
		MaxH: size.H,
	}
}

func (c Constraints) Clamp(size Size) Size {
	size.W = clamp(size.W, c.MinW, c.maxWidth())
	size.H = clamp(size.H, c.MinH, c.maxHeight())
	return size
}

func (c Constraints) Deflate(in Insets) Constraints {
	deflated := Constraints{
		MinW: max(0, c.MinW-in.Left-in.Right),
		MaxW: max(0, c.maxWidth()-in.Left-in.Right),
		MinH: max(0, c.MinH-in.Top-in.Bottom),
		MaxH: max(0, c.maxHeight()-in.Top-in.Bottom),
	}
	if deflated.MaxW == 0 && c.MaxW == 0 {
		deflated.MaxW = 0
	}
	if deflated.MaxH == 0 && c.MaxH == 0 {
		deflated.MaxH = 0
	}
	return deflated
}

func (c Constraints) maxWidth() float64 {
	if c.MaxW <= 0 {
		return math.MaxFloat64
	}
	return c.MaxW
}

func (c Constraints) maxHeight() float64 {
	if c.MaxH <= 0 {
		return math.MaxFloat64
	}
	return c.MaxH
}

type HitTarget struct {
	ID      string
	Rect    Rect
	Enabled bool
}

type Theme struct {
	Background    color.Color
	Backdrop      color.Color
	ModalFill     color.Color
	ModalStroke   color.Color
	PanelFill     color.Color
	PanelStroke   color.Color
	SectionFill   color.Color
	SectionStroke color.Color
	Title         color.Color
	Body          color.Color
	Muted         color.Color
	Error         color.Color
	ButtonFill    color.Color
	ButtonStroke  color.Color
	ButtonText    color.Color
	ActiveFill    color.Color
	ActiveStroke  color.Color
	DisabledFill  color.Color
	DisabledText  color.Color
	InputFill     color.Color
	InputStroke   color.Color
	InputFocus    color.Color
	ProgressTrack color.Color
	ProgressFill  color.Color
	WarningFill   color.Color
	WarningStroke color.Color
	GraphFill     color.Color
	GraphStroke   color.Color
	GraphGrid     color.Color
	GraphLine     color.Color
	AccentText    color.Color
}

func DefaultTheme() Theme {
	return DarkTheme()
}

func DarkTheme() Theme {
	// Omarchy's default Tokyo Night palette. Keep the colours semantic here so
	// every native surface (launcher, overlays, settings, and diagnostics)
	// shares the same restrained desktop feel.
	return Theme{
		Background:    color.RGBA{R: 26, G: 27, B: 38, A: 255}, // #1a1b26
		Backdrop:      color.RGBA{R: 14, G: 14, B: 20, A: 190},
		ModalFill:     color.RGBA{R: 19, G: 20, B: 28, A: 250},  // #13141c
		ModalStroke:   color.RGBA{R: 65, G: 72, B: 104, A: 220}, // #414868
		PanelFill:     color.RGBA{R: 36, G: 40, B: 59, A: 246},  // #24283b
		PanelStroke:   color.RGBA{R: 65, G: 72, B: 104, A: 210},
		SectionFill:   color.RGBA{R: 31, G: 35, B: 51, A: 250},
		SectionStroke: color.RGBA{R: 65, G: 72, B: 104, A: 190},
		Title:         color.RGBA{R: 192, G: 202, B: 245, A: 255}, // #c0caf5
		Body:          color.RGBA{R: 169, G: 177, B: 214, A: 255}, // #a9b1d6
		Muted:         color.RGBA{R: 86, G: 95, B: 137, A: 255},   // #565f89
		Error:         color.RGBA{R: 247, G: 118, B: 142, A: 255}, // #f7768e
		ButtonFill:    color.RGBA{R: 31, G: 35, B: 51, A: 255},
		ButtonStroke:  color.RGBA{R: 65, G: 72, B: 104, A: 255},
		ButtonText:    color.RGBA{R: 180, G: 190, B: 230, A: 255}, // #b4bee6
		ActiveFill:    color.RGBA{R: 41, G: 46, B: 66, A: 255},    // #292e42
		ActiveStroke:  color.RGBA{R: 122, G: 162, B: 247, A: 255}, // #7aa2f7
		DisabledFill:  color.RGBA{R: 19, G: 20, B: 28, A: 235},
		DisabledText:  color.RGBA{R: 65, G: 72, B: 104, A: 255},
		InputFill:     color.RGBA{R: 14, G: 14, B: 20, A: 255},
		InputStroke:   color.RGBA{R: 65, G: 72, B: 104, A: 255},
		InputFocus:    color.RGBA{R: 122, G: 162, B: 247, A: 255},
		ProgressTrack: color.RGBA{R: 19, G: 20, B: 28, A: 255},
		ProgressFill:  color.RGBA{R: 122, G: 162, B: 247, A: 255},
		WarningFill:   color.RGBA{R: 63, G: 49, B: 38, A: 255},
		WarningStroke: color.RGBA{R: 224, G: 175, B: 104, A: 255}, // #e0af68
		GraphFill:     color.RGBA{R: 19, G: 20, B: 28, A: 245},
		GraphStroke:   color.RGBA{R: 65, G: 72, B: 104, A: 220},
		GraphGrid:     color.RGBA{R: 65, G: 72, B: 104, A: 155},
		GraphLine:     color.RGBA{R: 125, G: 207, B: 255, A: 255}, // #7dcfff
		AccentText:    color.RGBA{R: 122, G: 162, B: 247, A: 255},
	}
}

func LightTheme() Theme {
	return Theme{
		Background:    color.RGBA{R: 236, G: 242, B: 248, A: 255},
		Backdrop:      color.RGBA{A: 120},
		ModalFill:     color.RGBA{R: 246, G: 250, B: 255, A: 248},
		ModalStroke:   color.RGBA{R: 93, G: 112, B: 138, A: 210},
		PanelFill:     color.RGBA{R: 233, G: 241, B: 249, A: 248},
		PanelStroke:   color.RGBA{R: 100, G: 120, B: 147, A: 205},
		SectionFill:   color.RGBA{R: 224, G: 234, B: 245, A: 250},
		SectionStroke: color.RGBA{R: 108, G: 128, B: 156, A: 205},
		Title:         color.RGBA{R: 23, G: 31, B: 44, A: 255},
		Body:          color.RGBA{R: 31, G: 41, B: 55, A: 255},
		Muted:         color.RGBA{R: 77, G: 93, B: 113, A: 255},
		Error:         color.RGBA{R: 176, G: 38, B: 54, A: 255},
		ButtonFill:    color.RGBA{R: 225, G: 235, B: 246, A: 255},
		ButtonStroke:  color.RGBA{R: 98, G: 118, B: 145, A: 205},
		ButtonText:    color.RGBA{R: 31, G: 41, B: 55, A: 255},
		ActiveFill:    color.RGBA{R: 48, G: 96, B: 168, A: 255},
		ActiveStroke:  color.RGBA{R: 27, G: 78, B: 168, A: 220},
		DisabledFill:  color.RGBA{R: 215, G: 224, B: 234, A: 230},
		DisabledText:  color.RGBA{R: 136, G: 146, B: 160, A: 255},
		InputFill:     color.RGBA{R: 249, G: 252, B: 255, A: 255},
		InputStroke:   color.RGBA{R: 98, G: 118, B: 145, A: 205},
		InputFocus:    color.RGBA{R: 59, G: 130, B: 246, A: 210},
		ProgressTrack: color.RGBA{R: 204, G: 216, B: 230, A: 255},
		ProgressFill:  color.RGBA{R: 59, G: 130, B: 246, A: 255},
		WarningFill:   color.RGBA{R: 255, G: 243, B: 214, A: 255},
		WarningStroke: color.RGBA{R: 217, G: 119, B: 6, A: 210},
		GraphFill:     color.RGBA{R: 227, G: 236, B: 246, A: 240},
		GraphStroke:   color.RGBA{R: 103, G: 123, B: 151, A: 200},
		GraphGrid:     color.RGBA{R: 174, G: 191, B: 211, A: 180},
		GraphLine:     color.RGBA{R: 37, G: 99, B: 235, A: 255},
		AccentText:    color.RGBA{R: 29, G: 78, B: 216, A: 255},
	}
}

type Context struct {
	Screen            *ebiten.Image
	Now               time.Time
	Theme             Theme
	MeasureText       func(value string, size float64) (float64, float64)
	MeasureWrapped    func(value string, width, size float64) float64
	DrawText          func(dst *ebiten.Image, value string, x, y, size float64, clr color.Color)
	DrawWrappedText   func(dst *ebiten.Image, value string, x, y, width, size float64, clr color.Color) float64
	RegisterHitTarget func(HitTarget)
	Runtime           *Runtime
	OnAction          func(id string)
}

func (c *Context) FillRect(r Rect, clr color.Color) {
	vector.FillRect(c.Screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), clr, false)
}

func (c *Context) StrokeRect(r Rect, width float64, clr color.Color) {
	if width <= 0 || r.W <= 0 || r.H <= 0 {
		return
	}
	strokeW := min(width, min(r.W, r.H))
	c.FillRect(Rect{X: r.X, Y: r.Y, W: r.W, H: strokeW}, clr)
	if r.H > strokeW {
		c.FillRect(Rect{X: r.X, Y: r.Bottom() - strokeW, W: r.W, H: strokeW}, clr)
	}
	if r.H > strokeW*2 {
		sideH := r.H - strokeW*2
		c.FillRect(Rect{X: r.X, Y: r.Y + strokeW, W: strokeW, H: sideH}, clr)
		if r.W > strokeW {
			c.FillRect(Rect{X: r.Right() - strokeW, Y: r.Y + strokeW, W: strokeW, H: sideH}, clr)
		}
	}
}

func (c *Context) StrokeLine(from, to Point, width float64, clr color.Color) {
	vector.StrokeLine(c.Screen, float32(from.X), float32(from.Y), float32(to.X), float32(to.Y), float32(width), clr, false)
}

func (c *Context) FillCircle(center Point, radius float64, clr color.Color) {
	if radius <= 0 {
		return
	}
	vector.FillCircle(c.Screen, float32(center.X), float32(center.Y), float32(radius), clr, false)
}

func (c *Context) FillRoundedRect(r Rect, radius float64, clr color.Color) {
	if r.W <= 0 || r.H <= 0 {
		return
	}
	radius = clamp(radius, 0, min(r.W/2, r.H/2))
	if radius <= 0 {
		c.FillRect(r, clr)
		return
	}
	c.FillRect(Rect{X: r.X + radius, Y: r.Y, W: r.W - radius*2, H: r.H}, clr)
	c.FillRect(Rect{X: r.X, Y: r.Y + radius, W: radius, H: r.H - radius*2}, clr)
	c.FillRect(Rect{X: r.Right() - radius, Y: r.Y + radius, W: radius, H: r.H - radius*2}, clr)
	c.FillCircle(Point{X: r.X + radius, Y: r.Y + radius}, radius, clr)
	c.FillCircle(Point{X: r.Right() - radius, Y: r.Y + radius}, radius, clr)
	c.FillCircle(Point{X: r.X + radius, Y: r.Bottom() - radius}, radius, clr)
	c.FillCircle(Point{X: r.Right() - radius, Y: r.Bottom() - radius}, radius, clr)
}

func (c *Context) FillStrokedRoundedRect(r Rect, width, radius float64, stroke, fill color.Color) {
	if width <= 0 || r.W <= 0 || r.H <= 0 {
		c.FillRoundedRect(r, radius, fill)
		return
	}
	width = min(width, min(r.W/2, r.H/2))
	c.FillRoundedRect(r, radius, stroke)
	inner := r.Inset(Insets{Top: width, Right: width, Bottom: width, Left: width})
	if inner.W <= 0 || inner.H <= 0 {
		return
	}
	c.FillRoundedRect(inner, max(0, radius-width), fill)
}

func (c *Context) AddHit(id string, r Rect, enabled bool) {
	if c.Runtime != nil {
		control := Control{ID: id, Rect: r, Enabled: enabled}
		if c.OnAction != nil && id != "" {
			control.OnClick = func(PointerEvent) {
				c.OnAction(id)
			}
		}
		c.Runtime.Register(control)
	}
	if c.RegisterHitTarget == nil || id == "" {
		return
	}
	c.RegisterHitTarget(HitTarget{ID: id, Rect: r, Enabled: enabled})
}

type Element interface {
	Measure(ctx *Context, constraints Constraints) Size
	Draw(ctx *Context, bounds Rect)
}

type fill struct {
	clr color.Color
}

func Fill(clr color.Color) Element {
	return fill{clr: clr}
}

func (f fill) Measure(_ *Context, constraints Constraints) Size {
	return constraints.Clamp(Size{})
}

func (f fill) Draw(ctx *Context, bounds Rect) {
	ctx.FillRect(bounds, f.clr)
}

func clamp(value, low, high float64) float64 {
	if high < low {
		high = low
	}
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
