package tracker

import (
	"encoding/json"
	"image"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type Tracker struct {
	items []Item

	sheetDisabled *ebiten.Image
	sheetEnabled  *ebiten.Image
}

func New(path string) (*Tracker, error) {
	items, err := loadItems(path)
	if err != nil {
		return nil, err
	}
	log.Printf("loaded %d items", len(items))

	sheetDisabled, _, err := ebitenutil.NewImageFromFile("assets/items-disabled.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	sheetEnabled, _, err := ebitenutil.NewImageFromFile("assets/items.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	return &Tracker{
		items:         items,
		sheetDisabled: sheetDisabled,
		sheetEnabled:  sheetEnabled,
	}, nil
}

// getItemIndex returns the index of the item at the given position or -1 if
// there is no item under the given pixel.
func (tracker *Tracker) getItemIndex(x, y int) int {
	for k := range tracker.items {
		if (image.Point{x, y}).In(tracker.items[k].Rect()) {
			return k
		}
	}

	return -1
}

// Upgrade upgrades the item under the given point.
func (tracker *Tracker) Upgrade(x, y int) {
	i := tracker.getItemIndex(x, y)
	if i < 0 {
		return
	}

	tracker.items[i].Upgrade()
}

// Downgrade downgrades the item under the given point.
func (tracker *Tracker) Downgrade(x, y int) {
	i := tracker.getItemIndex(x, y)
	if i < 0 {
		return
	}

	tracker.items[i].Downgrade()
}

func (tracker *Tracker) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}

	drawState := func(state bool, sheet *ebiten.Image) {
		for k := range tracker.items {
			if tracker.items[k].Enabled != state {
				continue
			}

			op.GeoM.Reset()
			op.GeoM.Translate(
				float64(tracker.items[k].X),
				float64(tracker.items[k].Y),
			)

			if err := screen.DrawImage(
				sheet.SubImage(tracker.items[k].SheetRect()).(*ebiten.Image),
				&op,
			); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Do two loops to avoid texture switches.
	drawState(false, tracker.sheetDisabled)
	drawState(true, tracker.sheetEnabled)
}

func loadItems(path string) ([]Item, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []Item
	dec := json.NewDecoder(f)
	if err := dec.Decode(&items); err != nil {
		return nil, err
	}

	return items, nil
}
