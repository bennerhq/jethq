package ui

import "testing"

func TestRuntimeHandlePointerClick(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{}
	runtime.BeginFrame()
	clicked := 0
	runtime.Register(Control{
		ID:      "button",
		Rect:    Rect{X: 10, Y: 20, W: 80, H: 24},
		Enabled: true,
		OnClick: func(PointerEvent) {
			clicked++
		},
	})

	point := Point{X: 30, Y: 30}
	if !runtime.HandlePointer(point, true, true, false) {
		t.Fatal("expected press to be consumed")
	}
	if !runtime.HandlePointer(point, false, false, true) {
		t.Fatal("expected release to be consumed")
	}
	if clicked != 1 {
		t.Fatalf("clicked = %d, want 1", clicked)
	}
}

func TestRuntimeHandlePointerClickWithoutID(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{}
	runtime.BeginFrame()
	clicked := 0
	runtime.Register(Control{
		Rect:    Rect{X: 10, Y: 20, W: 80, H: 24},
		Enabled: true,
		OnClick: func(PointerEvent) {
			clicked++
		},
	})

	point := Point{X: 30, Y: 30}
	runtime.HandlePointer(point, true, true, false)
	runtime.HandlePointer(point, false, false, true)
	if clicked != 1 {
		t.Fatalf("clicked = %d, want 1", clicked)
	}
}

func TestRuntimeHandlePointerUsesTopmostControl(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{}
	runtime.BeginFrame()
	clicked := ""
	runtime.Register(Control{
		ID:      "bottom",
		Rect:    Rect{X: 0, Y: 0, W: 100, H: 100},
		Enabled: true,
		OnClick: func(PointerEvent) {
			clicked = "bottom"
		},
	})
	runtime.Register(Control{
		ID:      "top",
		Rect:    Rect{X: 0, Y: 0, W: 100, H: 100},
		Enabled: true,
		OnClick: func(PointerEvent) {
			clicked = "top"
		},
	})

	point := Point{X: 30, Y: 30}
	runtime.HandlePointer(point, true, true, false)
	runtime.HandlePointer(point, false, false, true)
	if clicked != "top" {
		t.Fatalf("clicked = %q, want top", clicked)
	}
}

func TestRuntimeBoundsReturnsLatestRegisteredRect(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{}
	runtime.BeginFrame()
	runtime.Register(Control{ID: "field", Rect: Rect{X: 10, Y: 20, W: 30, H: 40}})
	runtime.Register(Control{ID: "field", Rect: Rect{X: 50, Y: 60, W: 70, H: 80}})

	got, ok := runtime.Bounds("field")
	if !ok {
		t.Fatal("expected bounds lookup to succeed")
	}
	want := (Rect{X: 50, Y: 60, W: 70, H: 80})
	if got != want {
		t.Fatalf("Bounds(field) = %+v, want %+v", got, want)
	}
}

func TestRuntimeWasClickedTracksClickByID(t *testing.T) {
	t.Parallel()

	runtime := &Runtime{}
	runtime.BeginFrame()
	runtime.Register(Control{
		ID:      "button",
		Rect:    Rect{X: 0, Y: 0, W: 40, H: 20},
		Enabled: true,
		OnClick: func(PointerEvent) {},
	})

	point := Point{X: 10, Y: 10}
	runtime.HandlePointer(point, true, true, false)
	runtime.HandlePointer(point, false, false, true)
	if !runtime.WasClicked("button") {
		t.Fatal("expected runtime to record button click")
	}

	runtime.BeginFrame()
	if runtime.WasClicked("button") {
		t.Fatal("expected click state to clear at next frame")
	}
}
