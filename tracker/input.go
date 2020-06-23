package tracker

import (
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
)

type kbInput struct {
	state             inputState
	activeKPZone      int // visually tied to a keypad number
	downgradeNextItem bool
}

type inputState int

const (
	inputStateIdle inputState = iota

	// Asking for a coarse KP zone
	inputStateItemKPZoneInput

	// Asking for an item inside a KP zone
	inputStateItemInput
)

func (tracker *Tracker) kbInputStateIs(v inputState) bool {
	return tracker.input.state == v
}

func (tracker *Tracker) kbInputStateIsAny(states ...inputState) bool {
	for _, v := range states {
		if tracker.input.state == v {
			return true
		}
	}

	return false
}

func (input *kbInput) reset() {
	*input = kbInput{}
}

func (tracker *Tracker) Input(input []rune) {
	if len(input) == 0 {
		return
	}

	for _, r := range input {
		tracker.inputAction(runeToAction(r))
	}
}

func (tracker *Tracker) inputAction(a action) {
	// Ensure we can _always_ leave using KP0
	if a == actionStartItemInput && !tracker.kbInputStateIs(inputStateIdle) {
		tracker.input.reset()
		return
	}

	switch tracker.input.state {
	case inputStateIdle:
		switch a {
		case actionIgnore:
			return
		case actionStartItemInput:
			tracker.input.state = inputStateItemKPZoneInput
		case actionDowngradeNext:
			tracker.input.state = inputStateItemKPZoneInput
			tracker.input.downgradeNextItem = !tracker.input.downgradeNextItem
		case actionTopLeft, actionTop, actionTopRight,
			actionLeft, actionMiddle, actionRight,
			actionBottomLeft, actionBottom, actionBottomRight:
			tracker.input.activeKPZone = actionToKPZone(a)
			tracker.input.state = inputStateItemInput
		case actionRedo:
			tracker.redo()
		case actionUndo:
			tracker.undo()
		}

	case inputStateItemKPZoneInput:
		switch a {
		case actionDowngradeNext:
			tracker.input.downgradeNextItem = !tracker.input.downgradeNextItem
		case actionTopLeft, actionTop, actionTopRight,
			actionLeft, actionMiddle, actionRight,
			actionBottomLeft, actionBottom, actionBottomRight:
			tracker.input.activeKPZone = actionToKPZone(a)
			tracker.input.state = inputStateItemInput
		default:
			// Reset on wrong input so we can start typing the correct "code" right away.
			tracker.input.reset()
		}

	case inputStateItemInput:
		if a == actionDowngradeNext {
			tracker.input.downgradeNextItem = !tracker.input.downgradeNextItem
			return
		}

		if err := tracker.inputKPZoneItem(tracker.input.activeKPZone, actionToKPZone(a)); err != nil {
			log.Printf("warning: %s", err)
			// Reset on wrong input so we can start typing the correct "code" right away.
			tracker.input.reset()
		}
	}
}

// inputKPZoneItem triggers an upgrade (or downgrade) of an item selected using
// first its rough then fine position on the tracker using the numpad.
func (tracker *Tracker) inputKPZoneItem(zoneKP, itemKP int) error {
	index, err := tracker.GetZoneItemIndex(zoneKP, itemKP)
	if err != nil {
		return err
	}

	tracker.changeItem(index, !tracker.input.downgradeNextItem)
	tracker.input.reset()

	return nil
}

func actionToKPZone(a action) int {
	switch a {
	case actionTopLeft:
		return 7
	case actionTop:
		return 8
	case actionTopRight:
		return 9
	case actionLeft:
		return 4
	case actionMiddle:
		return 5
	case actionRight:
		return 6
	case actionBottomLeft:
		return 1
	case actionBottom:
		return 2
	case actionBottomRight:
		return 3
	default:
		return -1
	}
}

// DEBUG
func (tracker *Tracker) drawInputState(screen *ebiten.Image) {
	pos := tracker.Origin.Add(image.Point{10, 15 + 9*gridSize})
	if tracker.kbInputStateIsAny(inputStateItemInput, inputStateItemKPZoneInput) {
		str := "+"
		if tracker.input.downgradeNextItem {
			str = "-"
		}

		text.Draw(screen, str, tracker.fontSmall, pos.X, pos.Y, color.White)
	}
}

type action int

const (
	actionIgnore action = iota
	actionStartItemInput
	actionDowngradeNext

	actionUndo
	actionRedo

	actionTopLeft
	actionTop
	actionTopRight
	actionLeft
	actionMiddle
	actionRight
	actionBottomLeft
	actionBottom
	actionBottomRight
)

// runeToAction is the keyboard "binds" part, as we handle text input and not
// keys we already are qwerty/azerty compatible but can't distinguish the main
// keyboard from keypad.
func runeToAction(r rune) action {
	switch r {
	case '0':
		return actionStartItemInput
	case '.':
		return actionDowngradeNext
	case '-':
		return actionUndo
	case '+':
		return actionRedo

	case '7':
		return actionTopLeft
	case '8':
		return actionTop
	case '9':
		return actionTopRight
	case '4':
		return actionLeft
	case '5':
		return actionMiddle
	case '6':
		return actionRight
	case '1':
		return actionBottomLeft
	case '2':
		return actionBottom
	case '3':
		return actionBottomRight
	default:
		return actionIgnore
	}
}
