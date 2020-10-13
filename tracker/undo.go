package tracker

import (
	"log"
	"strings"
)

// undoStackEntry represents an action (upgrade/downgrade) that happened on an item.
type undoStackEntry struct {
	HintText          string
	HintType          hintType
	ItemIndex         int
	IsHint, IsUpgrade bool
}

func (tracker *Tracker) appendHintToUndoStack(t hintType, str string) {
	// If we were back in time, discard and replace history.
	if len(tracker.redoStack) > 0 {
		tracker.redoStack = nil
	}

	tracker.undoStack = append(tracker.undoStack, undoStackEntry{
		IsHint:   true,
		HintType: t,
		HintText: str,
	})
}

func (tracker *Tracker) appendToUndoStack(itemIndex int, isUpgrade bool) {
	// If we were back in time, discard and replace history.
	if len(tracker.redoStack) > 0 {
		tracker.redoStack = nil
	}

	tracker.undoStack = append(tracker.undoStack, undoStackEntry{
		ItemIndex: itemIndex,
		IsUpgrade: isUpgrade,
	})
}
func (tracker *Tracker) undo() {
	if len(tracker.undoStack) == 0 {
		log.Printf("no action to undo")
		return
	}

	entry := tracker.undoStack[len(tracker.undoStack)-1]
	tracker.undoStack = tracker.undoStack[:len(tracker.undoStack)-1]
	tracker.redoStack = append(tracker.redoStack, entry)

	if entry.IsHint {
		switch entry.HintType {
		case hintTypeWOTH:
			tracker.undoWOTH(entry)
		case hintTypeBarren:
			tracker.barrens = tracker.barrens[:len(tracker.barrens)-1]
		case hintTypeSometimes:
			tracker.sometimes = tracker.sometimes[:len(tracker.sometimes)-1]
		case hintTypeAlways:
			index, _ := tracker.parseAlways(entry.HintText)
			tracker.setAlways(index, "")
		}
		return
	}

	if entry.IsUpgrade {
		tracker.items[entry.ItemIndex].Downgrade()
	} else {
		tracker.items[entry.ItemIndex].Upgrade()
	}
}

func (tracker *Tracker) undoWOTH(entry undoStackEntry) {
	// Clear last added double woth or last woth
	double := entry.HintText + doubleWOTHMarker
	for i := len(tracker.woths) - 1; i >= 0; i-- {
		if tracker.woths[i] == double {
			tracker.woths[i] = strings.TrimSuffix(tracker.woths[i], doubleWOTHMarker)
			return
		}
	}

	if len(tracker.woths) == 0 {
		return
	}

	tracker.woths = tracker.woths[:len(tracker.woths)-1]
}

func (tracker *Tracker) redo() {
	if len(tracker.redoStack) == 0 {
		log.Printf("no action to redo")
		return
	}

	entry := tracker.redoStack[len(tracker.redoStack)-1]
	tracker.redoStack = tracker.redoStack[:len(tracker.redoStack)-1]
	tracker.undoStack = append(tracker.undoStack, entry)

	if entry.IsHint {
		switch entry.HintType {
		case hintTypeWOTH:
			tracker.AddWOTH(entry.HintText)
		case hintTypeBarren:
			tracker.AddBarren(entry.HintText)
		case hintTypeSometimes:
			tracker.AddSometimes(entry.HintText)
		case hintTypeAlways:
			tracker.AddAlways(entry.HintText)
		}
		return
	}

	if entry.IsUpgrade {
		tracker.items[entry.ItemIndex].Upgrade()
	} else {
		tracker.items[entry.ItemIndex].Downgrade()
	}
}
