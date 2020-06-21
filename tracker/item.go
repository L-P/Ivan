package tracker

import (
	"image"
)

type Item struct {
	Name           string
	X, Y           int // position over the background
	SheetX, SheetY int `json:",omitempty"` // origin in the spritesheet

	// Both can be set if they are of the same length.
	CapacityProgression []int  `json:",omitempty"`
	ItemProgression     []Item `json:",omitempty"`

	// Index of the current item/capacity and temple upgrade
	upgradeIndex, templeIndex int

	// For countable items.
	CountMax, CountStep, count int

	IsMedallion, IsSong, Enabled, IsMarked bool `json:",omitempty"`
}

// Capacity returns the currently selected capacity of the item or -1 if it has
// no capacity to display.
func (item Item) Capacity() int {
	if !item.HasCapacity() {
		return -1
	}

	// HACK, use the index as a direct count.
	if item.IsCountable() {
		return item.upgradeIndex
	}

	return item.CapacityProgression[item.upgradeIndex]
}

const (
	gridSize         = 42
	itemSpriteWidth  = 34
	itemSpriteHeight = itemSpriteWidth

	// Item X/Y origin is 0, 0 adjusted by these margins.
	marginTop  = (gridSize - itemSpriteHeight) / 2
	marginLeft = (gridSize - itemSpriteWidth) / 2
)

// Rect returns the position of the item relative to the background origin.
func (item Item) Rect() image.Rectangle {
	return image.Rect(
		marginLeft+item.X,
		marginTop+item.Y,
		marginLeft+item.X+itemSpriteWidth,
		marginTop+item.Y+itemSpriteHeight,
	)
}

// SheetRect returns the position of the item sprite on the spritesheet.
func (item Item) SheetRect() image.Rectangle {
	x, y := item.SheetX, item.SheetY

	if len(item.ItemProgression) > 0 {
		if item.Enabled {
			x = item.ItemProgression[item.upgradeIndex].SheetX
			y = item.ItemProgression[item.upgradeIndex].SheetY
		} else {
			x = item.ItemProgression[len(item.ItemProgression)-1].SheetX
			y = item.ItemProgression[len(item.ItemProgression)-1].SheetY
		}
	}

	return image.Rect(
		x, y,
		x+itemSpriteWidth, y+itemSpriteHeight,
	)
}

// Upgrade upgrades the item to the next capacity or item upgrade (or both).
// If the item was disabled it does not upgrade it but enables it.
func (item *Item) Upgrade() {
	if !item.Enabled {
		item.Enabled = true
		return
	}

	if item.IsCountable() {
		item.countUp()
		return
	}

	var max int
	if len(item.ItemProgression) > 0 {
		max = len(item.ItemProgression)
	} else if len(item.CapacityProgression) > 0 {
		max = len(item.CapacityProgression)
	}

	if max == 0 || ((item.upgradeIndex + 1) >= max) {
		return // not upgradable, skip
	}

	item.upgradeIndex = (item.upgradeIndex + 1) % max
}

func (item *Item) Toggle() {
	item.Enabled = !item.Enabled
}

// Downgrades downgrades the item to the previous upgrade.
// It does not wrap around once the disabled state is reached.
func (item *Item) Downgrade() {
	if !item.Enabled {
		return
	}

	if item.IsCountable() {
		item.countDown()
		return
	}

	var max int
	if len(item.ItemProgression) > 0 {
		max = len(item.ItemProgression)
	} else if len(item.CapacityProgression) > 0 {
		max = len(item.CapacityProgression)
	}

	if max == 0 {
		item.Enabled = false
		return
	}

	if (item.upgradeIndex - 1) < 0 {
		item.Enabled = false
		return
	}

	item.upgradeIndex = (item.upgradeIndex - 1) % max
}

func (item *Item) countDown() {
	item.count -= item.CountStep
	if item.count < 0 {
		item.count = 0
		item.Enabled = false
	}
}

func (item *Item) countUp() {
	item.count += item.CountStep
	if item.count > item.CountMax {
		item.count = item.CountMax
	}
}

func (item *Item) Count() int {
	return item.count
}

func (item *Item) HasCapacity() bool {
	return len(item.CapacityProgression) > 0
}

func (item *Item) IsCountable() bool {
	return item.Name == "Golden Skulltulas" || item.Name == "Heart Piece"
}

func (item *Item) ToggleMark() {
	item.IsMarked = !item.IsMarked
}
