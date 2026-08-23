package ui

type PointerEvent struct {
	Point Point
	Local Point
	Rect  Rect
}

type Control struct {
	ID        string
	Rect      Rect
	Enabled   bool
	OnClick   func(PointerEvent)
	OnPress   func(PointerEvent)
	OnDrag    func(PointerEvent)
	OnRelease func(PointerEvent)
}

type Runtime struct {
	controls     []Control
	active       Control
	hasActive    bool
	clicked      map[string]bool
	toggleValues map[string]bool
	sliderValues map[string]float64
}

func (r *Runtime) BeginFrame() {
	r.controls = r.controls[:0]
	clear(r.clicked)
}

func (r *Runtime) Register(control Control) {
	r.controls = append(r.controls, control)
}

func (r *Runtime) Bounds(id string) (Rect, bool) {
	for i := len(r.controls) - 1; i >= 0; i-- {
		if r.controls[i].ID == id {
			return r.controls[i].Rect, true
		}
	}
	return Rect{}, false
}

func (r *Runtime) WasClicked(id string) bool {
	if r == nil || id == "" {
		return false
	}
	return r.clicked[id]
}

func (r *Runtime) ToggleValue(id string, fallback bool) bool {
	if r == nil || id == "" {
		return fallback
	}
	if value, ok := r.toggleValues[id]; ok {
		return value
	}
	return fallback
}

func (r *Runtime) SetToggleValue(id string, value bool) {
	if r == nil || id == "" {
		return
	}
	r.ensureState()
	r.toggleValues[id] = value
}

func (r *Runtime) SliderValue(id string, fallback float64) float64 {
	if r == nil || id == "" {
		return fallback
	}
	if value, ok := r.sliderValues[id]; ok {
		return value
	}
	return fallback
}

func (r *Runtime) SetSliderValue(id string, value float64) {
	if r == nil || id == "" {
		return
	}
	r.ensureState()
	r.sliderValues[id] = value
}

func (r *Runtime) HandlePointer(point Point, pressed, justPressed, justReleased bool) bool {
	if r == nil {
		return false
	}
	consumed := false
	if justPressed {
		if control, ok := r.hit(point); ok {
			r.active = control
			r.hasActive = true
			consumed = true
			if control.OnPress != nil {
				control.OnPress(control.event(point))
			}
		} else {
			r.hasActive = false
			r.active = Control{}
		}
	}
	if pressed && r.hasActive {
		consumed = true
		if r.active.OnDrag != nil {
			r.active.OnDrag(r.active.event(point))
		}
	}
	if justReleased {
		if r.hasActive {
			consumed = true
			event := r.active.event(point)
			if r.active.OnRelease != nil {
				r.active.OnRelease(event)
			}
			if r.active.OnClick != nil && r.active.Rect.Contains(point) {
				r.noteClick(r.active.ID)
				r.active.OnClick(event)
			}
		}
		r.hasActive = false
		r.active = Control{}
	}
	return consumed
}

func (r *Runtime) noteClick(id string) {
	if r == nil || id == "" {
		return
	}
	r.ensureState()
	r.clicked[id] = true
}

func (r *Runtime) ensureState() {
	if r.clicked == nil {
		r.clicked = make(map[string]bool)
	}
	if r.toggleValues == nil {
		r.toggleValues = make(map[string]bool)
	}
	if r.sliderValues == nil {
		r.sliderValues = make(map[string]float64)
	}
}

func (r *Runtime) hit(point Point) (Control, bool) {
	for i := len(r.controls) - 1; i >= 0; i-- {
		control := r.controls[i]
		if !control.Enabled || !control.Rect.Contains(point) {
			continue
		}
		return control, true
	}
	return Control{}, false
}

func (c Control) event(point Point) PointerEvent {
	return PointerEvent{
		Point: point,
		Local: Point{X: point.X - c.Rect.X, Y: point.Y - c.Rect.Y},
		Rect:  c.Rect,
	}
}
