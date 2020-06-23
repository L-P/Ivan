package tracker

import "log"

// undoStackEntry represents an action (upgrade/downgrade) that happened on an item.
type undoStackEntry struct {
	hintText          string
	hintType          hintType
	itemIndex         int
	isHint, isUpgrade bool
}

func (tracker *Tracker) appendHintToUndoStack(t hintType, str string) {
	// If we were back in time, discard and replace history.
	if len(tracker.redoStack) > 0 {
		tracker.redoStack = nil
	}

	tracker.undoStack = append(tracker.undoStack, undoStackEntry{
		isHint:   true,
		hintType: t,
		hintText: str,
	})
}

func (tracker *Tracker) appendToUndoStack(itemIndex int, isUpgrade bool) {
	// If we were back in time, discard and replace history.
	if len(tracker.redoStack) > 0 {
		tracker.redoStack = nil
	}

	tracker.undoStack = append(tracker.undoStack, undoStackEntry{
		itemIndex: itemIndex,
		isUpgrade: isUpgrade,
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

	if entry.isHint {
		switch entry.hintType {
		case hintTypeWOTH:
			tracker.woths = tracker.woths[:len(tracker.woths)-1]
		case hintTypeBarren:
			tracker.barrens = tracker.barrens[:len(tracker.barrens)-1]
		case hintTypeSometimes:
			tracker.sometimes = tracker.sometimes[:len(tracker.sometimes)-1]
		case hintTypeAlways:
			index, _ := parseAlways(entry.hintText)
			tracker.setAlways(index, "")
		}
		return
	}

	if entry.isUpgrade {
		tracker.items[entry.itemIndex].Downgrade()
	} else {
		tracker.items[entry.itemIndex].Upgrade()
	}
}

func (tracker *Tracker) redo() {
	if len(tracker.redoStack) == 0 {
		log.Printf("no action to redo")
		return
	}

	entry := tracker.redoStack[len(tracker.redoStack)-1]
	tracker.redoStack = tracker.redoStack[:len(tracker.redoStack)-1]
	tracker.undoStack = append(tracker.undoStack, entry)

	if entry.isHint {
		switch entry.hintType {
		case hintTypeWOTH:
			tracker.AddWOTH(entry.hintText)
		case hintTypeBarren:
			tracker.AddBarren(entry.hintText)
		case hintTypeSometimes:
			tracker.AddSometimes(entry.hintText)
		case hintTypeAlways:
			tracker.AddAlways(entry.hintText)
		}
		return
	}

	if entry.isUpgrade {
		tracker.items[entry.itemIndex].Upgrade()
	} else {
		tracker.items[entry.itemIndex].Downgrade()
	}
}
