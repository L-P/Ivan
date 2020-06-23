package tracker

import (
	"image"
	"image/color"
	"log"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/text"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func (tracker *Tracker) AddWOTH(str string) bool {
	if len(tracker.woths) >= 5 {
		log.Printf("warning: WotHs maxed out at 5")
		return false
	}

	tracker.woths = append(tracker.woths, str)
	return true
}

func (tracker *Tracker) AddBarren(str string) bool {
	if len(tracker.barrens) >= 3 {
		log.Printf("warning: barrens maxed out at 3")
		return false
	}

	tracker.barrens = append(tracker.barrens, str)
	return true
}

func (tracker *Tracker) AddSometimes(str string) bool {
	if len(tracker.sometimes) >= 5 {
		log.Printf("warning: sometimes maxed out at 5")
		return false
	}

	tracker.sometimes = append(tracker.sometimes, str)
	return true
}

func (tracker *Tracker) AddAlways(str string) bool {
	index, item := parseAlways(str)
	if index > -1 {
		tracker.setAlways(index, item)
		return true
	}

	log.Printf("warning: could not parse %s", str)
	return false
}

var alwaysLocations = []string{ // nolint:gochecknoglobals
	"Skull Mask", "Biggoron Sword",
	"30 Gold Skullutulas",
	"40 Gold Skullutulas",
	"50 Gold Skullutulas",
	"Ocarina of Time", "Frog 2",
}

func parseAlways(str string) (int, string) {
	parts := strings.SplitN(str, " ", 2)
	if len(parts) < 2 {
		parts = append(parts, "")
	}
	if parts[0] == "" {
		return -1, ""
	}

	matches := fuzzy.RankFindFold(parts[0], alwaysLocations)
	if len(matches) == 0 {
		return -1, ""
	}

	sort.Sort(matches)
	switch matches[0].Target {
	case "Skull Mask":
		return 0, parts[1]
	case "Biggoron Sword":
		return 1, parts[1]
	case "30 Gold Skullutulas":
		return 2, parts[1]
	case "40 Gold Skullutulas":
		return 3, parts[1]
	case "50 Gold Skullutulas":
		return 4, parts[1]
	case "Ocarina of Time":
		return 5, parts[1]
	case "Frog 2":
		return 6, parts[1]
	}

	return -1, ""
}

func (tracker *Tracker) setAlways(index int, str string) {
	if index < 0 || index >= len(tracker.always) {
		log.Printf(`bad index in setAlways(%d, "%s")`, index, str)
		return
	}

	tracker.always[index] = str
}

func (tracker *Tracker) drawHints(screen *ebiten.Image) {
	lineHeight := tracker.hintSize.Y / 10
	margins := image.Point{3, 15}

	pos := tracker.hintPos.Add(margins)
	for _, v := range tracker.woths {
		text.Draw(screen, v, tracker.fontSmall, pos.X, pos.Y, color.Black)
		pos.Y += lineHeight
	}

	pos = tracker.hintPos.Add(margins).Add(image.Point{tracker.hintSize.X / 2, 0})
	for _, v := range tracker.barrens {
		text.Draw(screen, v, tracker.fontSmall, pos.X, pos.Y, color.Black)
		pos.Y += lineHeight
	}

	pos = tracker.hintPos.Add(margins).Add(image.Point{tracker.hintSize.X / 2, 3 * lineHeight})
	for _, v := range tracker.sometimes {
		text.Draw(screen, v, tracker.fontSmall, pos.X, pos.Y, color.Black)
		pos.Y += lineHeight
	}

	pos = tracker.hintPos.Add(margins).Add(image.Point{22, 5 * lineHeight})
	for k, v := range tracker.always {
		if k == 5 {
			pos = pos.Add(image.Point{tracker.hintSize.X / 2, -2 * lineHeight})
		}

		text.Draw(screen, v, tracker.fontSmall, pos.X, pos.Y, color.Black)
		pos.Y += lineHeight
	}
}

func (tracker *Tracker) submitTextInput() {
	defer tracker.input.reset()

	if len(tracker.input.buf) == 0 {
		return
	}

	str := string(tracker.input.buf)
	if tracker.input.textInputFor == hintTypeWOTH ||
		tracker.input.textInputFor == hintTypeBarren {
		if match := tracker.matchLocation(str); match != "" {
			str = match
		}
	}

	var ok bool
	switch tracker.input.textInputFor {
	case hintTypeWOTH:
		ok = tracker.AddWOTH(str)
	case hintTypeBarren:
		ok = tracker.AddBarren(str)
	case hintTypeSometimes:
		ok = tracker.AddSometimes(str)
	case hintTypeAlways:
		ok = tracker.AddAlways(str)
	}

	if ok {
		tracker.appendHintToUndoStack(tracker.input.textInputFor, str)
	}
}
