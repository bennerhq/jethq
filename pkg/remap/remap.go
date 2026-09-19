// Package remap matches local keyboard chords to remote keyboard chords.
package remap

import (
	"fmt"
	"sort"

	"github.com/bennerhq/jethq/pkg/input"
	"github.com/bennerhq/jethq/pkg/protocol/hidrpc"
)

// Rule is a locally stored, focused-window shortcut remap.
type Rule struct {
	ID          string        `json:"id"`
	Trigger     []input.Key   `json:"trigger"`
	Output      []input.Key   `json:"output,omitempty"` // legacy single-stroke output
	OutputSteps [][]input.Key `json:"output_steps,omitempty"`
	OutputText  string        `json:"output_text,omitempty"`
}

type Result struct {
	Consumed    bool
	OutputSteps [][]input.Key
	OutputText  string
}

type Engine struct {
	rules  []Rule
	active string
}

func New(rules []Rule) *Engine {
	e := &Engine{}
	e.SetRules(rules)
	return e
}

func (e *Engine) SetRules(rules []Rule) {
	e.rules = make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if len(rule.OutputSteps) == 0 && len(rule.Output) > 0 {
			rule.OutputSteps = [][]input.Key{rule.Output}
		}
		if len(rule.Trigger) == 0 || (len(rule.OutputSteps) == 0 && rule.OutputText == "") {
			continue
		}
		rule.Trigger = normalized(rule.Trigger)
		for i := range rule.OutputSteps {
			rule.OutputSteps[i] = normalized(rule.OutputSteps[i])
		}
		e.rules = append(e.rules, rule)
	}
	e.active = ""
}

func (e *Engine) Reset() { e.active = "" }

func (e *Engine) Update(keys []input.Key) Result {
	keys = normalized(keys)
	if e.active != "" {
		if len(keys) == 0 {
			e.active = ""
		}
		return Result{Consumed: true}
	}
	for _, rule := range e.rules {
		if same(keys, rule.Trigger) {
			e.active = rule.ID
			return Result{Consumed: true, OutputSteps: append([][]input.Key(nil), rule.OutputSteps...), OutputText: rule.OutputText}
		}
	}
	for _, rule := range e.rules {
		if subset(keys, rule.Trigger) {
			return Result{Consumed: true}
		}
	}
	return Result{}
}

func normalized(keys []input.Key) []input.Key {
	seen := make(map[input.Key]bool, len(keys))
	result := make([]input.Key, 0, len(keys))
	for _, key := range keys {
		if key != input.KeyUnknown && !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func same(a, b []input.Key) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func subset(a, b []input.Key) bool {
	if len(a) == 0 || len(a) >= len(b) {
		return false
	}
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			i++
			j++
		case a[i] > b[j]:
			j++
		default:
			return false
		}
	}
	return i == len(a)
}

// MacroSteps makes a press-and-release HID macro for a remote shortcut.
func MacroSteps(chords [][]input.Key) ([]hidrpc.KeyboardMacroStep, error) {
	if len(chords) == 0 {
		return nil, fmt.Errorf("remote shortcut is empty")
	}
	steps := make([]hidrpc.KeyboardMacroStep, 0, len(chords)*2)
	for _, chord := range chords {
		var press hidrpc.KeyboardMacroStep
		for _, key := range normalized(chord) {
			hid, ok := input.KeyToHID(key)
			if !ok {
				return nil, fmt.Errorf("unsupported key %s", key.String())
			}
			if hid >= 224 && hid <= 231 {
				press.Modifier |= 1 << (hid - 224)
				continue
			}
			for i := range press.Keys {
				if press.Keys[i] == 0 {
					press.Keys[i] = hid
					break
				}
				if i == len(press.Keys)-1 {
					return nil, fmt.Errorf("remote shortcut has more than six non-modifier keys")
				}
			}
		}
		press.Delay = 20
		steps = append(steps, press, hidrpc.KeyboardMacroStep{Delay: 20})
	}
	return steps, nil
}
