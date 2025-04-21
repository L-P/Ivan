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
	UpgradeIndex, DungeonIndex int

	// For countable items.
	CountMax, CountStep, Count int

	IsMedallion, IsSong, Enabled bool `json:",omitempty"`
}

// Capacity returns the currently selected capacity of the item or -1 if it has
// no capacity to display.
func (item *Item) Capacity() int {
	if !item.HasCapacity() {
		return -1
	}

	// HACK, use the index as a direct count.
	if item.IsCountable() {
		return item.UpgradeIndex
	}

	return item.CapacityProgression[item.UpgradeIndex]
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
func (item *Item) Rect() image.Rectangle {
	return image.Rect(
		marginLeft+item.X,
		marginTop+item.Y,
		marginLeft+item.X+itemSpriteWidth,
		marginTop+item.Y+itemSpriteHeight,
	)
}

// SheetRect returns the position of the item sprite on the spritesheet.
func (item *Item) SheetRect() image.Rectangle {
	x, y := item.SheetX, item.SheetY

	if len(item.ItemProgression) > 0 {
		if item.Enabled {
			x = item.ItemProgression[item.UpgradeIndex].SheetX
			y = item.ItemProgression[item.UpgradeIndex].SheetY
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
// It returns false if the item was not affected.
func (item *Item) Upgrade() bool {
	if !item.Enabled {
		item.Enabled = true
		return true
	}

	if item.IsCountable() {
		item.countUp()
		return true
	}

	var maxProgression int
	if len(item.ItemProgression) > 0 {
		maxProgression = len(item.ItemProgression)
	} else if len(item.CapacityProgression) > 0 {
		maxProgression = len(item.CapacityProgression)
	}

	if maxProgression == 0 || ((item.UpgradeIndex + 1) >= maxProgression) {
		return false // not upgradable, skip
	}

	item.UpgradeIndex = (item.UpgradeIndex + 1) % maxProgression
	return true
}

// Downgrades downgrades the item to the previous upgrade.
// It returns false if the item was not affected.
func (item *Item) Downgrade() bool {
	if !item.Enabled {
		return false
	}

	if item.IsCountable() {
		item.countDown()
		return true
	}

	var maxProgression int
	if len(item.ItemProgression) > 0 {
		maxProgression = len(item.ItemProgression)
	} else if len(item.CapacityProgression) > 0 {
		maxProgression = len(item.CapacityProgression)
	}

	if maxProgression == 0 || (item.UpgradeIndex-1) < 0 {
		item.Enabled = false
		return true
	}

	item.UpgradeIndex = (item.UpgradeIndex - 1) % maxProgression
	return true
}

func (item *Item) countDown() {
	item.Count -= item.CountStep
	if item.Count < 0 {
		item.Count = 0
		item.Enabled = false
	}
}

func (item *Item) countUp() {
	item.Count += item.CountStep
	if item.Count > item.CountMax {
		item.Count = item.CountMax
	}
}

func (item *Item) HasCapacity() bool {
	return len(item.CapacityProgression) > 0
}

func (item *Item) IsCountable() bool {
	return item.CountMax != 0
}

var dungeons = []string{
	"", "Free",
	"Deku", "DC", "Jabu",
	"Forest", "Fire", "Water",
	"Spirit", "Shdw",
}

// map the dungeons global with actual dungeon names.
func dungeonToDungeonIndex(str string) int {
	return map[string]int{
		"":                 0,
		"Free":             1,
		"Deku Tree":        2,
		"Dodongo's Cavern": 3,
		"Jabu Jabu":        4, //nolint:dupword
		"Forest Temple":    5,
		"Fire Temple":      6,
		"Water Temple":     7,
		"Spirit Temple":    8,
		"Shadow Temple":    9,
	}[str]
}

func (item *Item) SetDungeon(dungeon string) {
	item.DungeonIndex = dungeonToDungeonIndex(dungeon)
}

func (item *Item) CycleDungeon(up bool) {
	if up {
		item.DungeonIndex = (item.DungeonIndex + 1) % len(dungeons)
		return
	}

	if item.DungeonIndex-1 < 0 {
		item.DungeonIndex = len(dungeons) - 1
		return
	}

	item.DungeonIndex--
}

func (item *Item) DungeonText() string {
	return dungeons[item.DungeonIndex]
}
