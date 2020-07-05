package tracker

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type kbInput struct {
	state             inputState
	activeKPZone      int // visually tied to a keypad number
	downgradeNextItem bool

	buf          []rune // text input buffer
	textInputFor hintType
}

type hintType int

const (
	_ hintType = iota
	hintTypeWOTH
	hintTypeBarren
	hintTypeSometimes
	hintTypeAlways
)

type inputState int

const (
	inputStateIdle inputState = iota

	// Asking for a coarse KP zone
	inputStateItemKPZoneInput

	// Asking for an item inside a KP zone
	inputStateItemInput

	// Writing raw text for a fuzzy search
	inputStateTextInput
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

	if tracker.input.state == inputStateTextInput {
		tracker.input.buf = append(tracker.input.buf, input...)
		return
	}

	for _, r := range input {
		tracker.inputAction(tracker.runeToAction(r))
	}
}

func (tracker *Tracker) idleHandleAction(a action) {
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

	case actionStartWOTHInput:
		tracker.input.state = inputStateTextInput
		tracker.input.textInputFor = hintTypeWOTH
	case actionStartBarrenInput:
		tracker.input.state = inputStateTextInput
		tracker.input.textInputFor = hintTypeBarren
	case actionStartAlwaysHintInput:
		tracker.input.state = inputStateTextInput
		tracker.input.textInputFor = hintTypeAlways
	case actionStartSometimesHintInput:
		tracker.input.state = inputStateTextInput
		tracker.input.textInputFor = hintTypeSometimes

	case actionRedo:
		tracker.redo()
	case actionUndo:
		tracker.undo()
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
		tracker.idleHandleAction(a)

	case inputStateTextInput:
		switch a {
		case actionSubmit:
			tracker.submitTextInput()
		case actionCancel:
			tracker.cancelTextInput()
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
	pos := tracker.pos.Add(image.Point{10, 15 + 9*gridSize})
	var str string

	switch tracker.input.state {
	case inputStateItemInput, inputStateItemKPZoneInput:
		if tracker.input.downgradeNextItem {
			str = "-"
		} else {
			str = "+"
		}

	case inputStateTextInput:
		str = "> " + string(tracker.input.buf)
		if tracker.input.textInputFor == hintTypeWOTH ||
			tracker.input.textInputFor == hintTypeBarren {
			if match := tracker.matchLocation(string(tracker.input.buf)); match != "" {
				str += " (" + match + ")"
			}
		} else if tracker.input.textInputFor == hintTypeAlways {
			index, _ := parseAlways(string(tracker.input.buf))
			if index > -1 {
				str += fmt.Sprintf(` (%s)`, alwaysLocations[index])
			}
		}
	}

	if str == "" {
		return
	}

	text.Draw(screen, str, tracker.fontSmall, pos.X, pos.Y, color.White)
}

func (tracker *Tracker) matchLocation(str string) string {
	if str == "" { // this matches Market for some reason.
		return ""
	}

	// HACK: Force some established conventions.
	switch strings.ToLower(strings.Trim(str, " ")) {
	case "dc":
		str = "Dodongo's Cavern"
	case "gv":
		str = "Gerudo Valley"
	case "gy":
		str = "Graveyard"
	case "ogc":
		str = "Outside Ganon's Castle"
	}

	matches := fuzzy.RankFindFold(str, tracker.locations)
	if len(matches) == 0 {
		return ""
	}
	sort.Sort(matches)
	return matches[0].Target
}

type action string

const (
	actionIgnore         action = "Ignore"
	actionStartItemInput action = "StartItemInput"
	actionDowngradeNext  action = "DowngradeNext"

	actionStartWOTHInput          action = "StartWOTHInput"
	actionStartBarrenInput        action = "StartBarrenInput"
	actionStartAlwaysHintInput    action = "StartAlwaysHintInput"
	actionStartSometimesHintInput action = "StartSometimesHintInput"
	actionSubmit                  action = "Submit"
	actionCancel                  action = "Cancel"

	actionUndo action = "Undo"
	actionRedo action = "Redo"

	actionTopLeft     action = "TopLeft"
	actionTop         action = "Top"
	actionTopRight    action = "TopRight"
	actionLeft        action = "Left"
	actionMiddle      action = "Middle"
	actionRight       action = "Right"
	actionBottomLeft  action = "BottomLeft"
	actionBottom      action = "Bottom"
	actionBottomRight action = "BottomRight"
)

// runeToAction is the keyboard "binds" part, as we handle text input and not
// keys we already are qwerty/azerty compatible but can't distinguish the main
// keyboard from keypad.
func (tracker *Tracker) runeToAction(r rune) action {
	a, ok := tracker.binds[string([]rune{r})]
	if !ok {
		return actionIgnore
	}

	return action(a)
}

// Submit is called the user presses Enter.
func (tracker *Tracker) Submit() {
	if !tracker.kbInputStateIs(inputStateTextInput) {
		return
	}

	tracker.inputAction(actionSubmit)
}

// Submit is called the user presses Escape.
func (tracker *Tracker) Cancel() {
	if !tracker.kbInputStateIs(inputStateTextInput) {
		return
	}

	tracker.inputAction(actionCancel)
}

func (tracker *Tracker) Backspace() {
	if len(tracker.input.buf) == 0 {
		return
	}

	tracker.input.buf = tracker.input.buf[:len(tracker.input.buf)-1]
}

func (tracker *Tracker) cancelTextInput() {
	tracker.input.reset()
}

// EatInput returns true if the tracker should reserve all text inputs for itself.
func (tracker *Tracker) EatInput() bool {
	return tracker.kbInputStateIs(inputStateTextInput)
}
