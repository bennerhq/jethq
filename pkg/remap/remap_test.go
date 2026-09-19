package remap

import (
	"github.com/bennerhq/jethq/pkg/input"
	"testing"
)

func TestEngineConsumesPrefixesAndFiresOnce(t *testing.T) {
	e := New([]Rule{{ID: "launcher", Trigger: []input.Key{input.KeyControlLeft, input.KeyAltLeft, input.KeySpace}, Output: []input.Key{input.KeyMetaLeft, input.KeySpace}}})
	if got := e.Update([]input.Key{input.KeyControlLeft}); !got.Consumed || len(got.OutputSteps) != 0 {
		t.Fatalf("prefix = %+v", got)
	}
	got := e.Update([]input.Key{input.KeyControlLeft, input.KeyAltLeft, input.KeySpace})
	if !got.Consumed || len(got.OutputSteps) != 1 || len(got.OutputSteps[0]) != 2 {
		t.Fatalf("trigger = %+v", got)
	}
	if got = e.Update([]input.Key{input.KeyControlLeft, input.KeyAltLeft, input.KeySpace}); !got.Consumed || len(got.OutputSteps) != 0 {
		t.Fatalf("repeat = %+v", got)
	}
	e.Update(nil)
	if got = e.Update([]input.Key{input.KeyControlLeft, input.KeyAltLeft, input.KeySpace}); len(got.OutputSteps) != 1 || len(got.OutputSteps[0]) != 2 {
		t.Fatalf("second trigger = %+v", got)
	}
}

func TestEngineSendsUnicodeOutput(t *testing.T) {
	e := New([]Rule{{ID: "at-sign", Trigger: []input.Key{input.KeyControlLeft, input.KeyAltLeft, input.KeySpace}, OutputText: "@"}})
	got := e.Update([]input.Key{input.KeyControlLeft, input.KeyAltLeft, input.KeySpace})
	if !got.Consumed || got.OutputText != "@" || len(got.OutputSteps) != 0 {
		t.Fatalf("trigger = %+v", got)
	}
}
