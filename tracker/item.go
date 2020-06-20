package tracker

import "image"

type Item struct {
	Name string

	X, Y           int // position over the background
	SheetX, SheetY int `json:",omitempty"` // origin in the spritesheet

	// Both can be set if they are of the same length.
	CapacityProgression []int  `json:",omitempty"`
	ItemProgression     []Item `json:",omitempty"`

	// Index of the current item/capacity upgrade, 0 means disabled
	UpgradeIndex int  `json:"-"`
	Enabled      bool `json:"-"`
}

// Capacity returns the currently selected capacity of the item or -1 if it has
// no capacity to display.
func (item Item) Capacity() int {
	if len(item.CapacityProgression) == 0 {
		return -1
	}

	return item.CapacityProgression[item.UpgradeIndex]
}

const (
	itemSpriteWidth  = 34
	itemSpriteHeight = itemSpriteWidth
)

// Rect returns the position of the item relative to the background origin.
func (item Item) Rect() image.Rectangle {
	return image.Rect(
		item.X, item.Y,
		item.X+itemSpriteWidth, item.Y+itemSpriteHeight,
	)
}

// SheetRect returns the position of the item sprite on the spritesheet.
func (item Item) SheetRect() image.Rectangle {
	x, y := item.SheetX, item.SheetY

	if len(item.ItemProgression) > 0 {
		if item.Enabled {
			x = item.ItemProgression[item.UpgradeIndex].SheetX
			y = item.ItemProgression[item.UpgradeIndex].SheetY
		} else {
			x = item.ItemProgression[0].SheetX
			y = item.ItemProgression[0].SheetY
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

	var max int
	if len(item.ItemProgression) > 0 {
		max = len(item.ItemProgression)
	} else if len(item.CapacityProgression) > 0 {
		max = len(item.CapacityProgression)
	}

	if max == 0 || ((item.UpgradeIndex + 1) >= max) {
		return // not upgradable, skip
	}

	item.UpgradeIndex = (item.UpgradeIndex + 1) % max
}

// Downgrades downgrades the item to the previous upgrade.
// It does not wrap around once the disabled state is reached.
func (item *Item) Downgrade() {
	if !item.Enabled {
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

	if (item.UpgradeIndex - 1) < 0 {
		item.Enabled = false
		return
	}

	item.UpgradeIndex = (item.UpgradeIndex - 1) % max
}
