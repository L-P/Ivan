package tracker

import (
	"encoding/json"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

type Tracker struct {
	Origin image.Point

	items         []Item
	font          font.Face
	sheetDisabled *ebiten.Image
	sheetEnabled  *ebiten.Image
}

const capacityFontSize = 20

func New(path string) (*Tracker, error) {
	items, err := loadItems(path)
	if err != nil {
		return nil, err
	}

	fontBin, err := ioutil.ReadFile("assets/Inconsolata-Regular.ttf")
	if err != nil {
		return nil, err
	}

	ttf, err := truetype.Parse(fontBin)
	if err != nil {
		return nil, err
	}

	sheetDisabled, _, err := ebitenutil.NewImageFromFile("assets/items-disabled.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	sheetEnabled, _, err := ebitenutil.NewImageFromFile("assets/items.png", ebiten.FilterDefault)
	if err != nil {
		return nil, err
	}

	tracker := &Tracker{
		items:         items,
		sheetDisabled: sheetDisabled,
		sheetEnabled:  sheetEnabled,
		font: truetype.NewFace(ttf, &truetype.Options{
			Size:    capacityFontSize,
			Hinting: font.HintingFull,
		}),
	}

	return tracker, nil
}

// getItemIndexByPos returns the index of the item at the given position or -1 if
// there is no item under the given pixel.
func (tracker *Tracker) getItemIndexByPos(x, y int) int {
	for k := range tracker.items {
		if (image.Point{x, y}).In(tracker.items[k].Rect()) {
			return k
		}
	}

	return -1
}

// Upgrade upgrades the item under the given point.
func (tracker *Tracker) Upgrade(x, y int) {
	i := tracker.getItemIndexByPos(x, y)
	if i < 0 {
		return
	}

	if tracker.items[i].IsMedallion || tracker.items[i].IsSong {
		tracker.items[i].Toggle()
		return
	}

	tracker.items[i].Upgrade()
}

// Downgrade downgrades the item under the given point.
func (tracker *Tracker) Downgrade(x, y int) {
	i := tracker.getItemIndexByPos(x, y)
	if i < 0 {
		return
	}

	switch {
	case tracker.items[i].IsMedallion:
	case tracker.items[i].IsSong:
		tracker.items[i].ToggleMark()
	default:
		tracker.items[i].Downgrade()
	}
}

func (tracker *Tracker) Draw(screen *ebiten.Image) {
	op := ebiten.DrawImageOptions{}

	drawState := func(state bool, sheet *ebiten.Image) {
		for k := range tracker.items {
			if tracker.items[k].Enabled != state {
				continue
			}

			pos := tracker.items[k].Rect().Min.Add(tracker.Origin)
			op.GeoM.Reset()
			op.GeoM.Translate(float64(pos.X), float64(pos.Y))

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
	tracker.drawMarks(screen)
	tracker.drawCapacities(screen)
}

func (tracker *Tracker) drawMarks(screen *ebiten.Image) {
	for k := range tracker.items {
		if !tracker.items[k].IsMarked {
			continue
		}

		rect := tracker.items[k].Rect()
		x, y := rect.Max.X-3*marginLeft, rect.Min.Y+4*marginTop
		text.Draw(screen, "×", tracker.font, x, y, color.RGBA{0x2F, 0xE6, 0x46, 0xFF})
	}
}

func (tracker *Tracker) drawCapacities(screen *ebiten.Image) {
	for k := range tracker.items {
		var count int
		switch {
		case tracker.items[k].HasCapacity():
			count = tracker.items[k].Capacity()
		case tracker.items[k].IsCountable():
			count = tracker.items[k].Count()
		default:
			continue
		}

		if !tracker.items[k].Enabled {
			continue
		}

		rect := tracker.items[k].Rect()
		x, y := rect.Min.X, rect.Max.Y

		// HACK, display skull count on the right
		if tracker.items[k].Name == "Golden Skulltulas" {
			x, y = rect.Min.X+gridSize+marginLeft, rect.Min.Y+marginTop+(gridSize/2)
		}

		str := strconv.Itoa(count)
		text.Draw(screen, str, tracker.font, x, y, color.White)
	}
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
