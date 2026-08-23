package ui

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestTextInputRepeatUsesElapsedTime(t *testing.T) {
	t.Parallel()

	now := time.Unix(100, 0)
	state := TextInputState{
		repeatKey:    ebiten.KeyBackspace,
		repeatNextAt: now.Add(textInputRepeatDelay),
		now: func() time.Time {
			return now
		},
		keyPressed: func(key ebiten.Key) bool {
			return key == ebiten.KeyBackspace
		},
	}

	if state.repeatingKeyPressed(ebiten.KeyBackspace) {
		t.Fatal("expected no repeat before the repeat delay elapses")
	}

	now = now.Add(textInputRepeatDelay)
	if !state.repeatingKeyPressed(ebiten.KeyBackspace) {
		t.Fatal("expected repeat exactly at the repeat deadline")
	}

	next := state.repeatNextAt
	now = now.Add(textInputRepeatInterval / 2)
	if state.repeatingKeyPressed(ebiten.KeyBackspace) {
		t.Fatal("expected no repeat before the repeat interval elapses")
	}

	now = next
	if !state.repeatingKeyPressed(ebiten.KeyBackspace) {
		t.Fatal("expected repeat at the next interval deadline")
	}
}

func TestTextInputRepeatAdvancesFromLargeTimeGap(t *testing.T) {
	t.Parallel()

	now := time.Unix(100, 0)
	state := TextInputState{
		repeatKey:    ebiten.KeyDelete,
		repeatNextAt: now,
		now: func() time.Time {
			return now
		},
		keyPressed: func(key ebiten.Key) bool {
			return key == ebiten.KeyDelete
		},
	}

	if !state.repeatingKeyPressed(ebiten.KeyDelete) {
		t.Fatal("expected repeat when the next repeat deadline has passed")
	}
	if !state.repeatNextAt.After(now) {
		t.Fatal("expected next repeat deadline to move into the future")
	}
}
