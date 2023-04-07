package tracker

import (
	"log"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

const doubleWOTHMarker = "*"

func (tracker *Tracker) AddWOTH(str string) bool {
	tracker.woths = append(tracker.woths, str)
	return true
}

func (tracker *Tracker) AddBarren(str string) bool {
	tracker.barrens = append(tracker.barrens, str)
	return true
}

func (tracker *Tracker) AddSometimes(str string) bool {
	tracker.sometimes = append(tracker.sometimes, str)
	return true
}

func (tracker *Tracker) AddAlways(str string) bool {
	index, item := tracker.parseAlways(str)
	if index > -1 {
		tracker.setAlways(index, item)
		return true
	}

	log.Printf("warning: could not parse %s", str)
	return false
}

func (tracker *Tracker) getAlwaysLocations() []string {
	return []string{
		"Skull Mask",
		"Biggoron Sword",
		"Ocarina of Time",
		"Sheik at Kakariko",
		"Frogs 2",
		"30 Gold Skullutulas",
		"40 Gold Skullutulas",
		"50 Gold Skullutulas",
	}
}

func (tracker *Tracker) parseAlways(str string) (int, string) {
	parts := strings.SplitN(strings.Trim(str, " "), " ", 2)
	if len(parts) < 2 {
		parts = append(parts, "")
	}
	if parts[0] == "" {
		return -1, ""
	}

	matches := fuzzy.RankFindFold(parts[0], tracker.getAlwaysLocations())
	if len(matches) == 0 {
		return -1, ""
	}

	sort.Sort(matches)
	switch matches[0].Target {
	case "Skull Mask":
		return 0, parts[1]
	case "Biggoron Sword":
		return 1, parts[1]
	case "Ocarina of Time":
		return 2, parts[1]
	case "Sheik at Kakariko":
		return 3, parts[1]
	case "Frogs 2":
		return 4, parts[1]
	case "30 Gold Skullutulas":
		return 5, parts[1]
	case "40 Gold Skullutulas":
		return 6, parts[1]
	case "50 Gold Skullutulas":
		return 7, parts[1]
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

func (tracker *Tracker) submitTextInput() {
	defer tracker.input.reset()

	if len(tracker.input.buf) == 0 {
		return
	}

	str := string(tracker.input.buf)
	if tracker.input.textInputFor == hintTypeBarren {
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
